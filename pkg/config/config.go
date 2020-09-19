package config

// Config is the interface definition of generic config storage
type Config interface {
	ListEntries() map[string]string
}
