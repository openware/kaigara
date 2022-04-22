package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceSetGetSecrets(t *testing.T) {
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")
	scopes := []string{"private", "public", "secret"}
	deploymentID := "opendax_uat"
	appName := "peatio"

	// Initialize Vault SecretStore
	ss, err := NewService(vaultAddr, vaultToken, deploymentID)
	if err != nil {
		t.Fatal(err)
	}

	for _, scope := range scopes {
		err := ss.Read(appName, scope)
		assert.NoError(t, err)

		// Set Secret in each scope
		err = ss.SetEntry(appName, scope, "key_"+scope, "value_"+scope)
		assert.NoError(t, err)

		// Save Secrets from memory to Vault
		err = ss.Write(appName, scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after save
		secret, err := ss.GetEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)
		assert.Equal(t, "value_"+scope, secret.(string))

		// Delete Secret in each scope
		err = ss.DeleteEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)

		// Get and assert Secrets in each scope after the deletion
		secret, err = ss.GetEntry(appName, scope, "key_"+scope)
		assert.NoError(t, err)
		assert.Equal(t, nil, secret)
	}
}
