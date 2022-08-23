package config

import (
	"github.com/openware/kaigara/pkg/sql"
	"github.com/openware/pkg/ika"
)

var ConfPath = ""

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	Storage      string `yaml:"secret_store" env:"KAIGARA_STORAGE_DRIVER" env-default:"sql"`
	DeploymentID string `yaml:"deployment_id" env:"KAIGARA_DEPLOYMENT_ID" env-default:"opendax_uat"`
	AppNames     string `yaml:"app_names" env:"KAIGARA_APP_NAME"`
	Scopes       string `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public,private,secret"`

	VaultToken string `yaml:"vault_token" env:"KAIGARA_VAULT_TOKEN" env-default:"changeme"`
	VaultAddr  string `yaml:"vault_addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`

	EncryptMethod string `yaml:"encryption_method" env:"KAIGARA_ENCRYPTOR" env-default:"plaintext"`
	AesKey        string `yaml:"aes_key" env:"KAIGARA_ENCRYPTOR_AES_KEY" env-default:"changemechangeme"`

	KubeConfig string `yaml:"kubeconfig" env:"KUBECONFIG"`

	LogLevel int                `yaml:"log_level" env:"KAIGARA_LOG_LEVEL" env-default:"1"`
	RedisURL string             `yaml:"redis_url" env:"KAIGARA_REDIS_URL"`
	DBConfig sql.DatabaseConfig `yaml:"database"`
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
	conf := &KaigaraConfig{DBConfig: sql.DatabaseConfig{}}
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
			conf.DBConfig.Host = "127.0.0.1"
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
