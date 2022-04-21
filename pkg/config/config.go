package config

import (
	"github.com/openware/pkg/database"
)

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	Storage       string          `yaml:"secret-store" env:"KAIGARA_STORAGE_DRIVER" env-default:"vault"`
	VaultToken    string          `yaml:"vault-token" env:"KAIGARA_VAULT_TOKEN"`
	VaultAddr     string          `yaml:"vault-addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`
	IgnoreGlobal  bool            `yaml:"ignore-global" env:"KAIGARA_IGNORE_GLOBAL"`
	AppNames      string          `yaml:"vault-app-name" env:"KAIGARA_APP_NAME"`
	DeploymentID  string          `yaml:"deployment-id" env:"KAIGARA_DEPLOYMENT_ID"`
	Scopes        string          `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public,private,secret"`
	EncryptMethod string          `yaml:"encryption-method" env:"KAIGARA_ENCRYPTOR" env-default:"transit"`
	AesKey        string          `yaml:"aes-key" env:"KAIGARA_ENCRYPTOR_AES_KEY"`
	LogLevel      int             `yaml:"log-level" env:"KAIGARA_LOG_LEVEL" env-default:"1"`
	RedisURL      string          `yaml:"redis-url" env:"KAIGARA_REDIS_URL"`
	DBConfig      database.Config `yaml:"database"`
}

// Config is the interface definition of generic config storage
type Config interface {
	ListEntries() map[string]interface{}
}

// Env contains envrionment vars and file content and paths
type Env struct {
	Vars  []string
	Files map[string]*File
}

// File contains path and content of a file fetched from env by Kaigara
type File struct {
	Path    string
	Content string
}
