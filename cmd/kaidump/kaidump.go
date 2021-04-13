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
//	"gopkg.in/yaml.v3"
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
	filepath := flag.String("-a", "outputs.yaml", "Outputs file path to save secrets from vault")
	flag.Parse()
	fmt.Println("Outputs:", filepath)
	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")
	apps, err := secretStore.ListAppNames()
	if err != nil {
		panic(err)
	}
	if apps != nil {
		for app := range apps {
			fmt.Println("App name:", app)
		}
	}	
}
