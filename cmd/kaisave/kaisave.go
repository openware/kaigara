package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/pkg/ika"
	"gopkg.in/yaml.v3"
)

var cnf = &config.KaigaraConfig{}

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
	if err := ika.ReadConfig("", cnf); err != nil {
		panic(err)
	}

	secretStore, err := config.GetStorageService(cnf)
	if err != nil {
		panic(err)
	}

	for app, scopes := range secrets.Secrets {
		for scope, data := range scopes.Scopes {
			err := secretStore.Read(app, scope)
			if err != nil {
				panic(err)
			}

			for k, v := range data {
				fmt.Println("Setting", k)

				err := secretStore.SetEntry(app, scope, k, v)
				if err != nil {
					panic(err)
				}
			}

			err = secretStore.Write(app, scope)
			if err != nil {
				panic(err)
			}
		}
	}
}
