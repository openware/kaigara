package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// AESEncryptor implements Encryptor interface by using AES
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor instantiate an in memory encryption service
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("AES key length should be exactly 16, 24 or 32, actual length: %d", len(key))
	}

	return &AESEncryptor{
		key: key,
	}, nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (ae *AESEncryptor) Encrypt(plaintext, appName string) (string, error) {
	cipherBlock, err := aes.NewCipher(ae.key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Decrypt the given ciphertext and return the plaintext or an error
func (ae *AESEncryptor) Decrypt(ciphertext, appName string) (string, error) {
	encryptData, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	cipherBlock, err := aes.NewCipher(ae.key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(encryptData) < nonceSize {
		return "", err
	}

	nonce, cipherText := encryptData[:nonceSize], encryptData[nonceSize:]
	plainData, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainData), nil
}
