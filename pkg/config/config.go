package config

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/openware/kaigara/pkg/storage/sql"
	"github.com/openware/kaigara/pkg/storage/vault"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/database"
)

// KaigaraConfig contains cli options
type KaigaraConfig struct {
	SecretStore  string `yaml:"secret-store" env:"KAIGARA_SECRET_STORE" env-default:"vault"`
	VaultToken   string `yaml:"vault-token" env:"KAIGARA_VAULT_TOKEN"`
	VaultAddr    string `yaml:"vault-addr" env:"KAIGARA_VAULT_ADDR" env-default:"http://127.0.0.1:8200"`
	AppNames     string `yaml:"vault-app-name" env:"KAIGARA_APP_NAME"`
	DeploymentID string `yaml:"deployment-id" env:"KAIGARA_DEPLOYMENT_ID"`
	Scopes       string `yaml:"scopes" env:"KAIGARA_SCOPES" env-default:"public,private,secret"`
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

func GetStorageService(cnf *KaigaraConfig, db *database.Config) types.Storage {
	ss := cnf.SecretStore
	if ss == "vault" {
		return vault.NewService(cnf.VaultAddr, cnf.VaultToken, cnf.DeploymentID)
	} else if ss == "sql" {
		svc, err := sql.NewStorageService(cnf.DeploymentID, db)
		if err != nil {
			panic(err)
		}
		return svc
	} else {
		panic("SecretStore does not support: " + ss)
	}
}

// BuildCmdEnv reads secrets from all secretStores and scopes passed to it and loads them into an Env and returns a *Env
func BuildCmdEnv(appNames []string, store types.Storage, currentEnv, scopes []string) *Env {
	env := &Env{
		Vars:  []string{},
		Files: map[string]*File{},
	}

	for _, v := range currentEnv {
		if !strings.HasPrefix(v, "KAIGARA_") {
			env.Vars = append(env.Vars, v)
		}
	}

	for _, appName := range append([]string{"global"}, appNames...) {
		for _, scope := range scopes {
			err := store.Read(appName, scope)
			if err != nil {
				panic(err)
			}

			secrets, err := store.GetEntries(appName, scope)
			if err != nil {
				panic(err)
			}

			for k, v := range secrets {
				// Avoid trying to put maps and slices into env
				if _, ok := v.(map[string]interface{}); ok {
					continue
				}

				if _, ok := v.([]interface{}); ok {
					continue
				}

				var val string

				// Handle bool and json.Number
				if tmp, ok := v.(bool); ok {
					val = strconv.FormatBool(tmp)
				}

				if tmp, ok := v.(json.Number); ok {
					val = string(tmp)
				}

				// Skip if the var can't be asserted to string
				if val == "" {
					if tmp, ok := v.(string); ok {
						val = tmp
					} else {
						continue
					}
				}

				m := kfile.FindStringSubmatch(k)

				if m == nil {
					env.Vars = append(env.Vars, strings.ToUpper(k)+"="+val)
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
