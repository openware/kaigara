package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestAESEncryptorWrongKey(t *testing.T) {
	_, err := NewAESEncryptor([]byte("too_short"))
	require.Error(t, err)
}

func TestAESEncryptor(t *testing.T) {
	s, err := NewAESEncryptor([]byte("1234567890123456"))
	require.NoError(t, err)

	cipher, err := s.Encrypt("bonjour", "")
	require.NoError(t, err)

	plain, err := s.Decrypt(cipher, "")
	require.NoError(t, err)
	assert.Equal(t, "bonjour", plain)
}
