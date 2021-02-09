package types

// SecretStore is used to store secrets
type SecretStore interface {
	LoadSecrets(scope string) error
	SetSecret(name, value string) error
	SetSecrets(data map[string]interface{}) error
	SaveSecrets(scope string) error
	GetSecret(name string) (interface{}, error)
	GetSecrets() (map[string]interface{}, error)
}
