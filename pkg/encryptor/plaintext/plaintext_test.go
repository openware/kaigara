package plaintext

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestPlaintextEncryptor(t *testing.T) {
	s := NewPlaintextEncryptor()
	require.NotEqual(t, s, nil)

	cipher, err := s.Encrypt("bonjour", "")
	require.NoError(t, err)
	assert.Equal(t, "bonjour", cipher)

	plain, err := s.Decrypt(cipher, "")
	require.NoError(t, err)
	assert.Equal(t, "bonjour", plain)
}
