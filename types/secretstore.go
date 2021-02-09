package types

// SecretStore is used to store secrets
type SecretStore interface {
	LoadSecrets(scope string) error
	SetSecret(name, value, scope string) error
	SetSecrets(data map[string]interface{}, scope string) error
	SaveSecrets(scope string) error
	GetSecret(name, scope string) (interface{}, error)
	GetSecrets(scope string) (map[string]interface{}, error)
}
