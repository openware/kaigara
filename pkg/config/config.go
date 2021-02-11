package config

import (
	"log"
	"regexp"
	"strings"

	"github.com/openware/kaigara/types"
)

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	SecretStore  string `yaml:"secret-store" env:"KAIGARA_SECRET_STORE" env-default:"vault"`
	VaultToken   string `yaml:"vault-token" env:"KAIGARA_VAULT_TOKEN"`
	VaultAddr    string `yaml:"vault-addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`
	AppName      string `yaml:"vault-app-name" env:"KAIGARA_APP_NAME"`
	DeploymentID string `yaml:"deployment-id" env:"KAIGARA_DEPLOYMENT_ID"`
	Scopes       string `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public"`
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

var kfile = regexp.MustCompile("(?i)^KFILE_(.*)_(PATH|CONTENT)$")

// BuildCmdEnv reads secrets from all secretStores and scopes passed to it and loads them into an Env and returns a *Env
func BuildCmdEnv(secretStores []types.SecretStore, currentEnv, scopes []string) *Env {
	env := &Env{
		Vars:  []string{},
		Files: map[string]*File{},
	}

	for _, v := range currentEnv {
		if !strings.HasPrefix(v, "KAIGARA_") {
			env.Vars = append(env.Vars, v)
		}
	}

	for _, secretStore := range secretStores {

		if secretStore == nil {
			continue
		}

		for _, scope := range scopes {
			err := secretStore.LoadSecrets(scope)
			if err != nil {
				panic(err)
			}

			secrets, err := secretStore.GetSecrets(scope)
			if err != nil {
				panic(err)
			}

			for k, v := range secrets {
				m := kfile.FindStringSubmatch(k)

				if m == nil {
					env.Vars = append(env.Vars, strings.ToUpper(k)+"="+v.(string))
					continue
				}
				name := strings.ToUpper(m[1])
				suffix := strings.ToUpper(m[2])

				f, ok := env.Files[name]
				if !ok {
					f = &File{}
					env.Files[name] = f
				}
				switch suffix {
				case "PATH":
					f.Path = v.(string)
				case "CONTENT":
					f.Content = v.(string)
				default:
					log.Printf("ERROR: Unexpected prefix in config key: %s", k)
				}
			}
		}
	}
	return env
}
