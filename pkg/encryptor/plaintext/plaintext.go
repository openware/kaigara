package plaintext

import (
	"fmt"
)

// PlaintextEncryptor implements Encryptor interface
type PlaintextEncryptor struct {
}

func NewPlaintextEncryptor() *PlaintextEncryptor {
	return &PlaintextEncryptor{}
}

// Encrypt implements plaintext encryption which will return just what you passed to it as argument
func (de *PlaintextEncryptor) Encrypt(plaintext, appName string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("encrypted value is empty")
	}

	return plaintext, nil
}

// Decrypt implements plaintext decryption which will return just what you passed to it as argument
func (de *PlaintextEncryptor) Decrypt(ciphertext, appName string) (string, error) {
	if ciphertext == "" {
		return "", fmt.Errorf("decrypted value is empty")
	}

	return ciphertext, nil
}
