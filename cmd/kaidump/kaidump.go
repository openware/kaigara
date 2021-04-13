package main

import (
	"flag"
	"fmt"
//	"io/ioutil"
	"os"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/vault"

	"github.com/openware/pkg/ika"
	"strings" // Needed to use Split
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
	filepath := flag.String("a", "outputs.yaml", "Outputs file path to save secrets from vault")
	flag.Parse()

	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")
	apps, err := secretStore.ListAppNames()
	if err != nil {
		panic(err)
	}
	scopes := os.Getenv("KAIGARA_SCOPES")
	if scopes == "" {
		panic("Undefined KAIGARA_SCOPES env")
	}
	scopesList := strings.Split(scopes, ",")
	if len(scopesList) <= 0 {
		panic("Wrong KAIGARA_SCOPES")
	}
	appMap := make(map[string]interface{})
	for _, app := range apps {
		scopeMap := make(map[string]interface{})
		for _, scope := range scopesList {
			err := secretStore.LoadSecrets(app, scope)
			if err != nil {
				panic(err)
			}		
			secrets, err := secretStore.GetSecrets(app, scope)
			if err != nil {
				panic(err)
			}
			scopeMap[scope] = secrets
		}
		appMap[app] = scopeMap
	}
	res, err := yaml.Marshal(&appMap)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--- secrets dump:\n%s\n\n", string(res))
	err = ioutil.WriteFile(*filepath, res, 0644)
	if err != nil {
		panic(err)
	}

	f.Close()
}
