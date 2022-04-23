package config

import (
	"github.com/openware/pkg/database"
	"github.com/openware/pkg/ika"
)

var ConfPath = ""

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	Storage string `yaml:"secret-store" env:"KAIGARA_STORAGE_DRIVER" env-default:"sql"`

	DeploymentID string `yaml:"deployment-id" env:"KAIGARA_DEPLOYMENT_ID" env-default:"opendax_uat"`
	AppNames     string `yaml:"vault-app-name" env:"KAIGARA_APP_NAME"`
	Scopes       string `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public,private,secret"`
	IgnoreGlobal bool   `yaml:"ignore-global" env:"KAIGARA_IGNORE_GLOBAL"`

	VaultToken string `yaml:"vault-token" env:"KAIGARA_VAULT_TOKEN" env-default:"changeme"`
	VaultAddr  string `yaml:"vault-addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`

	EncryptMethod string `yaml:"encryption-method" env:"KAIGARA_ENCRYPTOR" env-default:"plaintext"`
	AesKey        string `yaml:"aes-key" env:"KAIGARA_ENCRYPTOR_AES_KEY" env-default:"changemechangeme"`

	LogLevel int `yaml:"log-level" env:"KAIGARA_LOG_LEVEL" env-default:"1"`

	RedisURL string `yaml:"redis-url" env:"KAIGARA_REDIS_URL"`

	DBConfig database.Config `yaml:"database"`
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

func NewKaigaraConfig() (*KaigaraConfig, error) {
	conf := &KaigaraConfig{DBConfig: database.Config{}}
	if err := ika.ReadConfig(ConfPath, conf); err != nil {
		return nil, err
	}

	if conf.Storage == "sql" {
		if conf.DBConfig.Pool == 0 {
			conf.DBConfig.Pool = 1
		}

		if conf.DBConfig.Driver == "" {
			conf.DBConfig.Driver = "postgres"
		}

		if conf.DBConfig.Host == "" {
			conf.DBConfig.Host = conf.DBConfig.Driver
		}

		if conf.DBConfig.Port == "" {
			switch conf.DBConfig.Driver {
			case "mysql":
				conf.DBConfig.Port = "3306"
			case "postgres":
				conf.DBConfig.Port = "5432"
			}
		}

		if conf.DBConfig.User == "" {
			switch conf.DBConfig.Driver {
			case "mysql":
				conf.DBConfig.User = "root"
			case "postgres":
				conf.DBConfig.User = "postgres"
			}
		}

		if conf.DBConfig.Pass == "" {
			switch conf.DBConfig.Driver {
			case "mysql":
				conf.DBConfig.Pass = ""
			case "postgres":
				conf.DBConfig.Pass = "changeme"
			}
		}
	}

	return conf, nil
}
