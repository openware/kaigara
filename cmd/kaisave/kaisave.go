package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/pkg/ika"
	"gopkg.in/yaml.v3"
)

var cnf = &config.KaigaraConfig{}

func initConfig() {
	err := ika.ReadConfig("", cnf)
	if err != nil {
		panic(err)
	}
}

func getVaultService(appName string) *vault.Service {
	return vault.NewService(cnf.VaultAddr, cnf.VaultToken, cnf.DeploymentID)
}

func initLogStream() logstream.LogStream {
	url := os.Getenv("KAIGARA_REDIS_URL")
	return logstream.NewRedisClient(url)
}

// App contains a map of scopes(public, private, secret) with secrets to be loaded
type App struct {
	Scopes map[string]map[string]interface{} `yaml:"scopes"`
}

// SecretsFile contains secrets a map of Apps containing secrets to be loaded into the SecretStore
type SecretsFile struct {
	Secrets map[string]App `yaml:"secrets"`
}

func main() {
	// Parse flags
	filepath := flag.String("filepath", "secrets.yaml", "Path to the file containing secrets")
	flag.Parse()

	// Read the file
	fmt.Println("Reading the secrets file...")
	dat, err := ioutil.ReadFile(*filepath)
	if err != nil {
		panic(err)
	}

	// Parse YAML
	secrets := SecretsFile{}

	err = yaml.Unmarshal(dat, &secrets)
	if err != nil {
		panic(err)
	}

	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")

	for app, scopes := range secrets.Secrets {
		for scope, data := range scopes.Scopes {
			err := secretStore.LoadSecrets(app, scope)
			if err != nil {
				panic(err)
			}

			for k, v := range data {
				fmt.Println("Setting", k)

				err := secretStore.SetSecret(app, k, v, scope)
				if err != nil {
					panic(err)
				}
			}

			err = secretStore.SaveSecrets(app, scope)
			if err != nil {
				panic(err)
			}
		}
	}
}
