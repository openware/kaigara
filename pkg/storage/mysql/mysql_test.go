package mysql

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageServiceSetGetEntries(t *testing.T) {

	dsn := os.Getenv("KAIGARA_MYSQL_DSN")
	scopes := []string{"private", "public", "secret"}
	deploymentID := "opendax_uat"
	appName := "peatio"

	// Initialize StorageService
	ss, err := NewStorageService(dsn, deploymentID)
	assert.NoError(t, err)

	for _, scope := range scopes {
		err := ss.Read(appName, scope)
		assert.NoError(t, err)

		// Verify the initial datastore has been initialized properly
		data, err := ss.GetEntries(appName, scope)
		assert.NoError(t, err)
		assert.Equal(t, make(map[string]interface{}), data)

		// Set Entry in each scope
		err = ss.SetEntry(appName, scope, "key_"+scope, "value_"+scope)
		assert.NoError(t, err)

		// Save Entrys from memory to Vault
		err = ss.Write(appName, scope)
		assert.NoError(t, err)

		// Create a StorageService from scratch
		ssTmp, err := NewStorageService(dsn, deploymentID)
		assert.NoError(t, err)

		fmt.Printf("Loading ssTmp for %s", scope)
		err = ssTmp.Read(appName, scope)
		assert.NoError(t, err)

		// Get and assert Entries in each scope after save
		entry, err := ssTmp.GetEntry(appName, scope, "key_"+scope)
		fmt.Printf("entry: %s\n", entry)
		assert.NoError(t, err)
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

		ssTmp2, err := NewStorageService(dsn, deploymentID)
		assert.NoError(t, err)

		fmt.Printf("Loading ssTmp2 for %s\n", scope)
		err = ssTmp2.Read(appName, scope)
		assert.NoError(t, err)

		data, err = ssTmp2.GetEntries(appName, scope)
		assert.NoError(t, err)
		assert.Equal(t, make(map[string]interface{}), data)
	}
}
