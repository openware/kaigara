package types

// Storage is used to store data
type Storage interface {
	// Low level functions to retrieve or store all configuration entries
	Read(appName, scope string) error
	Write(appName, scope string) error

	// In memory actions
	SetEntry(appName, scope, name string, value interface{}) error
	SetEntries(appName, scope string, data map[string]interface{}) error
	GetEntry(appName, scope, name string) (interface{}, error)
	GetEntries(appName, scope string) (map[string]interface{}, error)
	ListEntries(appName, scope string) ([]string, error)
	DeleteEntry(appName, scope, name string) error
	ListAppNames() ([]string, error)

	// Get current version in memory
	GetCurrentVersion(appName, scope string) (int64, error)

	// Get latest version from the storage
	GetLatestVersion(appName, scope string) (int64, error)
}
