package sql

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/encryptor/aes"
	"github.com/openware/kaigara/pkg/encryptor/plaintext"
	"github.com/openware/kaigara/pkg/encryptor/transit"
	"github.com/openware/kaigara/pkg/encryptor/types"
)

var deploymentID = "opendax_uat"
var appNames = []string{"barong", "finex", "peatio"}
var scopes = []string{"private", "public", "secret"}
var configs map[string]config.DatabaseConfig
var encryptors map[string]types.Encryptor

type Config struct {
	Name     string                `yaml:"name"`
	DbConfig config.DatabaseConfig `yaml:"database"`
}

func TestMain(m *testing.M) {
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")

	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dat, err := ioutil.ReadFile(path + "/testdata/config.yml")
	if err != nil {
		panic(err)
	}

	cfgs := make(map[string]Config)
	err = yaml.Unmarshal(dat, &cfgs)
	if err != nil {
		panic(err)
	}

	configs = make(map[string]config.DatabaseConfig)
	for _, cfg := range cfgs {
		configs[cfg.Name] = cfg.DbConfig
	}

	aesEncrypt, err := aes.NewAESEncryptor([]byte("1234567890123456"))
	if err != nil {
		panic(err)
	}

	plainEncrypt := plaintext.NewPlaintextEncryptor()

	transitEncrypt, err := transit.NewVaultEncryptor(vaultAddr, vaultToken)
	if err != nil {
		panic(err)
	}

	encryptors = map[string]types.Encryptor{
		"aes":       aesEncrypt,
		"transit":   transitEncrypt,
		"plaintext": plainEncrypt,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func getEntriesReload(ss *Service, appName, scope string) (map[string]interface{}, error) {
	if err := ss.Read(appName, scope); err != nil {
		return nil, err
	}

	data, err := ss.GetEntries(appName, scope)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getEntryReload(ss *Service, appName, scope, name string) (interface{}, error) {
	if err := ss.Read(appName, scope); err != nil {
		return nil, err
	}

	entry, err := ss.GetEntry(appName, scope, name)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func setEntry(ss *Service, appName, scope, name, value string) error {
	if err := ss.SetEntry(appName, scope, name, value); err != nil {
		return err
	}

	// Save entries from memory to Vault
	if err := ss.Write(appName, scope); err != nil {
		return err
	}

	return nil
}

func clearStorage(conf config.DatabaseConfig) error {
	db, err := Connect(&conf)
	if err != nil {
		return err
	}

	tx := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Data{})
	return tx.Error
}

func TestSetEntry(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, conf := range configs {
			ss, err := NewService(deploymentID, &conf, encryptor, 1)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					init, err := getEntriesReload(ss, appName, scope)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, map[string]interface{}{"version": int64(0)}, init)

					name := "key_" + scope
					val := "value_" + scope
					if err := setEntry(ss, appName, scope, name, val); err != nil {
						t.Fatal(err)
					}

					// Verify the written data with new storage
					ssTmp, err := NewService(deploymentID, &conf, encryptor, 1)
					assert.NoError(t, err)

					result, err := getEntriesReload(ssTmp, appName, scope)
					if err != nil {
						t.Fatal(err)
					}

					delete(result, "version")
					assert.Equal(t, map[string]interface{}{"key_" + scope: "value_" + scope}, result)

					entry, err := getEntryReload(ssTmp, appName, scope, name)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, val, entry.(string))
				}
			}

			err = clearStorage(conf)
			assert.NoError(t, err)
		}
	}
}

func TestDeleteEntry(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, conf := range configs {
			ss, err := NewService(deploymentID, &conf, encryptor, 1)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					key := "key_" + scope
					val := "value_" + scope

					if _, err := getEntriesReload(ss, appName, scope); err != nil {
						t.Fatal(err)
					}

					if err := setEntry(ss, appName, scope, key, val); err != nil {
						t.Fatal(err)
					}

					// Delete Entry in each scope
					err = ss.DeleteEntry(appName, scope, key)
					assert.NoError(t, err)

					entry, err := ss.GetEntry(appName, scope, key)
					assert.NoError(t, err)
					assert.Equal(t, nil, entry)

					// Check that Write() will delete redundant data
					err = ss.Write(appName, scope)
					assert.NoError(t, err)

					ssTmp, err := NewService(deploymentID, &conf, encryptor, 1)
					assert.NoError(t, err)

					entry, err = getEntryReload(ssTmp, appName, scope, key)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, nil, entry)
				}
			}

			err = clearStorage(conf)
			assert.NoError(t, err)
		}
	}
}

func TestListAppNames(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, conf := range configs {
			ss, err := NewService(deploymentID, &conf, encryptor, 1)
			assert.NoError(t, err)

			for _, appName := range appNames {
				for _, scope := range scopes {
					key := "key_" + scope
					val := "value_" + scope

					if _, err := getEntriesReload(ss, appName, scope); err != nil {
						t.Fatal(err)
					}

					if err := setEntry(ss, appName, scope, key, val); err != nil {
						t.Fatal(err)
					}
				}
			}

			ssTmp, err := NewService(deploymentID, &conf, encryptor, 1)
			assert.NoError(t, err)

			apps, err := ssTmp.ListAppNames()
			sort.Strings(apps)
			assert.NoError(t, err)
			assert.Equal(t, appNames, apps)

			err = clearStorage(conf)
			assert.NoError(t, err)
		}
	}
}

func TestServiceSetGetEntriesIncreaseVersion(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, conf := range configs {
			ss, err := NewService(deploymentID, &conf, encryptor, 1)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					data, err := getEntriesReload(ss, appName, scope)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, map[string]interface{}{"version": int64(0)}, data)

					if err := setEntry(ss, appName, scope, "key_"+scope, "value_"+scope); err != nil {
						t.Fatal(err)
					}

					// Create a Service from scratch
					ssTmp, err := NewService(deploymentID, &conf, encryptor, 1)
					assert.NoError(t, err)

					// Get and assert Entries in each scope after save
					entry, err := getEntryReload(ssTmp, appName, scope, "key_"+scope)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, "value_"+scope, entry.(string))

					// Delete Entry in each scope
					err = ssTmp.DeleteEntry(appName, scope, "key_"+scope)
					assert.NoError(t, err)

					// Get and assert Entries in each scope after the deletion
					entry, err = ssTmp.GetEntry(appName, scope, "key_"+scope)
					assert.NoError(t, err)
					assert.Equal(t, nil, entry)

					// Check that Write() will delete redundant data
					err = ssTmp.Write(appName, scope)
					assert.NoError(t, err)

					ssTmp2, err := NewService(deploymentID, &conf, encryptor, 1)
					assert.NoError(t, err)

					data, err = getEntriesReload(ssTmp2, appName, scope)
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, map[string]interface{}{"version": int64(1)}, data)
				}
			}

			err = clearStorage(conf)
			assert.NoError(t, err)
		}
	}
}
