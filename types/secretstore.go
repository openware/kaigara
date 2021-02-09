package types

// SecretStore is used to store secrets
type SecretStore interface {
	LoadSecrets() error
	SetSecret(name, value string) error
	SetSecrets(data map[string]interface{}) error
	SaveSecrets() error
	GetSecret(name string) (interface{}, error)
	GetSecrets() (map[string]interface{}, error)
}