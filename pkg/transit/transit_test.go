package transit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestEncrypt(t *testing.T) {
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultAppName := os.Getenv("VAULT_APP_NAME")

	if vaultToken == "" || vaultAddr == "" || vaultAppName == "" {
		t.Skip()
		return
	}

	s := NewVaultEncryptor(vaultAddr, vaultToken, vaultAppName)
	cipher, err := s.Encrypt("bonjour")
	require.NoError(t, err)

	plain, err := s.Decrypt(cipher)
	require.NoError(t, err)
	assert.Equal(t, "bonjour", plain)
}
