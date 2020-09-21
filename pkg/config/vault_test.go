package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultConfigStoreListEntriesPresent(t *testing.T) {
	path := "test/vault_config_store"
	config := NewVaultConfig(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"), path)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"foo":  "bar",
			"fooz": "barz",
		},
	}

	_, err := config.client.Logical().Write("secret/data/"+path, data)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"foo":  "bar",
		"fooz": "barz",
	}, config.ListEntries())
}

func TestVaultConfigStoreListEntriesAbsent(t *testing.T) {
	path := "test/vault_config_store"
	config := NewVaultConfig(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"), path)
	_, err := config.client.Logical().Delete("secret/data/" + path)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, config.ListEntries())
}
