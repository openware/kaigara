package sql

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/ika"
	"github.com/openware/kaigara/pkg/aes"
	"github.com/openware/kaigara/pkg/plaintext"
	"github.com/openware/kaigara/pkg/transit"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	deploymentID string
	appNames     []string
	scopes       []string
	configs      map[string]database.Config
	encryptors   map[string]map[string]types.Encryptor
}

func (s *Suite) SetupSuite() {
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")

	s.deploymentID = "opendax_uat"
	s.appNames = []string{"barong", "finex", "peatio"}
	s.scopes = []string{"private", "public", "secret"}

	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	type Config struct {
		Name     string          `yaml:"name"`
		DbConfig database.Config `yaml:"database"`
	}
	cfgs := make(map[string]Config)
	ika.ReadConfig(path+"/config.yml", &cfgs)
	s.configs = make(map[string]database.Config)
	for _, cfg := range cfgs {
		s.configs[cfg.Name] = cfg.DbConfig
	}
	aesEncrypt, err := aes.NewAESEncryptor([]byte("1234567890123456"))
	if err != nil {
		panic(err)
	}

	plainEncrypt := plaintext.NewPlaintextEncryptor([]byte(""))
	s.encryptors = map[string]map[string]types.Encryptor{
		"aes": {
			"barong": aesEncrypt,
			"finex":  aesEncrypt,
			"peatio": aesEncrypt,
		},
		"transit": {
			"barong": transit.NewVaultEncryptor(vaultAddr, vaultToken, "barong"),
			"finex":  transit.NewVaultEncryptor(vaultAddr, vaultToken, "finex"),
			"peatio": transit.NewVaultEncryptor(vaultAddr, vaultToken, "peatio"),
		},
		"plaintext": {
			"barong": plainEncrypt,
			"finex":  plainEncrypt,
			"peatio": plainEncrypt,
		},
	}
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func getEntries(ss *StorageService, appName, scope string) map[string]interface{} {
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

func getEntry(ss *StorageService, appName, scope, name string) interface{} {
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
	// Set Entry in each scope
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

	tx := db.Debug().Unscoped().Where("1 = 1").Delete(&Data{})
	return tx.Error
}

func (s *Suite) AfterTest(_, _ string) {
	for _, cnf := range s.configs {
		err := clearStorage(cnf)
		assert.NoError(s.T(), err)
	}
}

func (s *Suite) TestSetEntry() {
	deploymentID := s.deploymentID
	t := s.T()

	for _, encryptor := range s.encryptors {
		for _, cnf := range s.configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor)
			assert.NoError(t, err)

			for _, scope := range s.scopes {
				for _, appName := range s.appNames {
					init := getEntries(ss, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(0)}, init)

					name := "key_" + scope
					val := "value_" + scope
					setEntry(ss, appName, scope, name, val)

					// Verify the written data with new storage
					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor)
					assert.NoError(t, err)

					result := getEntries(ssTmp, appName, scope)
					assert.NoError(t, err)
					assert.Equal(t, map[string]interface{}{"key_" + scope: "value_" + scope, "version": int64(0)}, result)

					entry := getEntry(ssTmp, appName, scope, name)
					assert.Equal(t, val, entry.(string))
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func (s *Suite) TestDeleteEntry() {
	deploymentID := s.deploymentID
	t := s.T()

	for _, encryptor := range s.encryptors {
		for _, cnf := range s.configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor)
			assert.NoError(t, err)

			for _, scope := range s.scopes {
				for _, appName := range s.appNames {
					key := "key_" + scope
					val := "value_" + scope

					getEntries(ss, appName, scope)

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

					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor)
					assert.NoError(t, err)

					entry = getEntry(ssTmp, appName, scope, key)
					assert.Equal(t, nil, entry)
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func (s *Suite) TestListAppNames() {
	deploymentID := s.deploymentID
	appNames := []string{"barong", "finex", "peatio"}
	t := s.T()

	for _, encryptor := range s.encryptors {
		for _, cnf := range s.configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor)
			assert.NoError(t, err)

			for _, appName := range appNames {
				for _, scope := range s.scopes {
					key := "key_" + scope
					val := "value_" + scope

					getEntries(ss, appName, scope)
					setEntry(ss, appName, scope, key, val)
				}
			}

			ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor)
			assert.NoError(t, err)

			apps, err := ssTmp.ListAppNames()
			assert.Equal(t, appNames, apps)

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}

func (s *Suite) TestStorageServiceSetGetEntriesIncreaseVersion() {
	deploymentID := s.deploymentID
	t := s.T()

	for _, encryptor := range s.encryptors {
		for _, cnf := range s.configs {
			ss, err := NewStorageService(deploymentID, &cnf, encryptor)
			assert.NoError(t, err)

			for _, scope := range s.scopes {
				for _, appName := range s.appNames {
					data := getEntries(ss, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(0)}, data)

					setEntry(ss, appName, scope, "key_"+scope, "value_"+scope)

					// Create a StorageService from scratch
					ssTmp, err := NewStorageService(deploymentID, &cnf, encryptor)
					assert.NoError(t, err)

					// Get and assert Entries in each scope after save
					entry := getEntry(ssTmp, appName, scope, "key_"+scope)
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

					ssTmp2, err := NewStorageService(deploymentID, &cnf, encryptor)
					assert.NoError(t, err)

					data = getEntries(ssTmp2, appName, scope)
					assert.Equal(t, map[string]interface{}{"version": int64(1)}, data)
				}
			}

			err = clearStorage(cnf)
			assert.NoError(t, err)
		}
	}
}
