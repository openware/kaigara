package types

// Storage is used to store data
type Storage interface {
	LoadSecrets(appName, scope string) error
	SaveSecrets(appName, scope string) error
	SetSecret(appName, name string, value interface{}, scope string) error
	SetSecrets(appName string, data map[string]interface{}, scope string) error
	GetSecret(appName, name, scope string) (interface{}, error)
	GetSecrets(appName, scope string) (map[string]interface{}, error)
	ListSecrets(appName, scope string) ([]string, error)
	DeleteSecret(appName, name, scope string) error
	ListAppNames() ([]string, error)
	GetCurrentVersion(appName, scope string) (int64, error)
	GetLatestVersion(appName, scope string) (int64, error)
}
