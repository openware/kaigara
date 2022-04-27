package transit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestEncrypt(t *testing.T) {
	vaultToken := os.Getenv("KAIGARA_VAULT_TOKEN")
	vaultAddr := os.Getenv("KAIGARA_VAULT_ADDR")
	vaultAppName := "finex"

	if vaultToken == "" || vaultAddr == "" {
		t.Fatal("vault settings are not set")
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
