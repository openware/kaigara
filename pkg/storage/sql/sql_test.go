package sql

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/openware/kaigara/pkg/encryptor/aes"
	"github.com/openware/kaigara/pkg/encryptor/plaintext"
	"github.com/openware/kaigara/pkg/encryptor/transit"
	"github.com/openware/kaigara/pkg/encryptor/types"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var deploymentID = "opendax_uat"
var appNames = []string{"barong", "finex", "peatio"}
var scopes = []string{"private", "public", "secret"}
var configs map[string]database.Config
var encryptors map[string]types.Encryptor

type Config struct {
	Name     string          `yaml:"name"`
	DbConfig database.Config `yaml:"database"`
}

func TestMain(m *testing.M) {
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")

	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dat, err := ioutil.ReadFile(path + "/config.yml")
	if err != nil {
		panic(err)
	}

	cfgs := make(map[string]Config)
	err = yaml.Unmarshal(dat, &cfgs)
	if err != nil {
		panic(err)
	}

	configs = make(map[string]database.Config)
	for _, cfg := range cfgs {
		configs[cfg.Name] = cfg.DbConfig
	}
	aesEncrypt, err := aes.NewAESEncryptor([]byte("1234567890123456"))
	if err != nil {
		panic(err)
	}

	plainEncrypt := plaintext.NewPlaintextEncryptor()
	encryptors = map[string]types.Encryptor{
		"aes":       aesEncrypt,
		"transit":   transit.NewVaultEncryptor(vaultAddr, vaultToken),
		"plaintext": plainEncrypt,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func getEntriesReload(ss *StorageService, appName, scope string) map[string]interface{} {
	err := ss.Read(appName, scope)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loading ss for %s\n", scope)
	data, err := ss.GetEntries(appName, scope)
	if err != nil {
		panic(err)
	}
	return data
}

func getEntryReload(ss *StorageService, appName, scope, name string) interface{} {
	err := ss.Read(appName, scope)
	if err != nil {
		panic(err)
	}

	entry, err := ss.GetEntry(appName, scope, name)
	fmt.Printf("entry: %s\n", entry)
	if err != nil {
		panic(err)
	}
	return entry
}

func setEntry(ss *StorageService, appName, scope, name, value string) {
	err := ss.SetEntry(appName, scope, name, value)
	if err != nil {
		panic(err)
	}

	// Save Entrys from memory to Vault
	err = ss.Write(appName, scope)
	if err != nil {
		panic(err)
	}
}

func clearStorage(cnf database.Config) error {
	db, err := database.Connect(&cnf)
	if err != nil {
		return err
	}

	tx := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&Data{})
	return tx.Error
}

func TestSetEntry(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, cnf := range configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					init := getEntriesReload(ss, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(0)}, init)

					name := "key_" + scope
					val := "value_" + scope
					setEntry(ss, appName, scope, name, val)

					// Verify the written data with new storage
					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
					assert.NoError(t, err)

					result := getEntriesReload(ssTmp, appName, scope)
					assert.NoError(t, err)
					assert.Equal(t, map[string]interface{}{"key_" + scope: "value_" + scope, "version": int64(0)}, result)

					entry := getEntryReload(ssTmp, appName, scope, name)
					assert.Equal(t, val, entry.(string))
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func TestDeleteEntry(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, cnf := range configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					key := "key_" + scope
					val := "value_" + scope

					getEntriesReload(ss, appName, scope)

					setEntry(ss, appName, scope, key, val)

					// Delete Entry in each scope
					err = ss.DeleteEntry(appName, scope, key)
					assert.NoError(t, err)

					entry, err := ss.GetEntry(appName, scope, key)
					assert.NoError(t, err)
					assert.Equal(t, nil, entry)

					// Check that Write() will delete redundant data
					err = ss.Write(appName, scope)
					assert.NoError(t, err)

					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
					assert.NoError(t, err)

					entry = getEntryReload(ssTmp, appName, scope, key)
					assert.Equal(t, nil, entry)
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func TestListAppNames(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, cnf := range configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
			assert.NoError(t, err)

			for _, appName := range appNames {
				for _, scope := range scopes {
					key := "key_" + scope
					val := "value_" + scope

					getEntriesReload(ss, appName, scope)
					setEntry(ss, appName, scope, key, val)
				}
			}

			ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
			assert.NoError(t, err)

			apps, err := ssTmp.ListAppNames()
			sort.Strings(apps)
			assert.NoError(t, err)
			assert.Equal(t, appNames, apps)

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func TestStorageServiceSetGetEntriesIncreaseVersion(t *testing.T) {
	for _, encryptor := range encryptors {
		for _, cnf := range configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
			assert.NoError(t, err)

			for _, scope := range scopes {
				for _, appName := range appNames {
					data := getEntriesReload(ss, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(0)}, data)

					setEntry(ss, appName, scope, "key_"+scope, "value_"+scope)

					// Create a StorageService from scratch
					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
					assert.NoError(t, err)

					// Get and assert Entries in each scope after save
					entry := getEntryReload(ssTmp, appName, scope, "key_"+scope)
					fmt.Printf("entry: %s\n", entry)
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

					ssTmp2, err := NewStorageService(deploymentID, &cnf, encryptor, logger.Silent)
					assert.NoError(t, err)

					data = getEntriesReload(ssTmp2, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(1)}, data)
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}
