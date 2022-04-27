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

	s, err := NewVaultEncryptor(vaultAddr, vaultToken)
	if err != nil {
		t.Fatal(err)
	}

	cipher, err := s.Encrypt("bonjour", vaultAppName)
	require.NoError(t, err)

	plain, err := s.Decrypt(cipher, vaultAppName)
	require.NoError(t, err)
	assert.Equal(t, "bonjour", plain)
}
