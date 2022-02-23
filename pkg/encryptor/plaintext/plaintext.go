package plaintext

import (
	"fmt"
)

// PlaintextEncryptor implements Encryptor interface
type PlaintextEncryptor struct {
	key []byte
}

func NewPlaintextEncryptor(key []byte) *PlaintextEncryptor {
	return &PlaintextEncryptor{
		key: key,
	}
}

// Encrypt implements plaintext encryption which will return just what you passed to it as argument
func (de *PlaintextEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("Encrypted value is empty!")
	}

	return plaintext, nil
}

// Decrypt implements plaintext decryption which will return just what you passed to it as argument
func (de *PlaintextEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("Decrypted value is empty!")
	}

	return ciphertext, nil
}
