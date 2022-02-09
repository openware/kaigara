package types

// Encryptor is used to encrypt/decrypt data for storage drivers
type Encryptor interface {
	Encrypt(ciphertext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}
