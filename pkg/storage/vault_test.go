package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultServiceSetGetSecrets(t *testing.T) {
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")
	scopes := []string{"private", "public", "secret"}
	deploymentID := "opendax_uat"
	appName := "peatio"

	// Initialize Vault SecretStore
	secretStore, err := NewVaultService(vaultAddr, vaultToken, deploymentID)
	if err != nil {
		t.Fatal(err)
	}

	for _, scope := range scopes {
		err := secretStore.Read(appName, scope)
		assert.NoError(t, err)

		// Set Secret in each scope
		err = secretStore.SetEntry(appName, scope, "key_"+scope, "value_"+scope)
		assert.NoError(t, err)

		// Save Secrets from memory to Vault
		err = secretStore.Write(appName, scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after save
		secret, err := secretStore.GetEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)
		assert.Equal(t, "value_"+scope, secret.(string))

		// Delete Secret in each scope
		err = secretStore.DeleteEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after the deletion
		secret, err = secretStore.GetEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)
		assert.Equal(t, nil, secret)
	}
}
