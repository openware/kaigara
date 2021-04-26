package vault

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
	secretStore := NewService(vaultAddr, vaultToken, deploymentID)

	for _, scope := range scopes {
		err := secretStore.LoadSecrets(appName, scope)
		assert.NoError(t, err)

		// Set Secret in each scope
		err = secretStore.SetSecret(appName, "key_"+scope, "value_"+scope, scope)
		assert.NoError(t, err)

		// Save Secrets from memory to Vault
		err = secretStore.SaveSecrets(appName, scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after save
		secret, err := secretStore.GetSecret(appName, "key_"+scope, scope)
		assert.NoError(t, err)
		assert.Equal(t, "value_"+scope, secret.(string))

		// Delete Secret in each scope
		err = secretStore.DeleteSecret(appName, "key_"+scope, scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after the deletion
		secret, err = secretStore.GetSecret(appName, "key_"+scope, scope)
		assert.NoError(t, err)
		assert.Equal(t, nil, secret)
	}
}
