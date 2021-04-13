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
	filepath := flag.String("a", "outputs.yaml", "Outputs file path to save secrets from vault")
	appName := flag.String("appname", "global", "App name")
	scope := flag.String("scope", "public", "Scope name")
	key := flag.String("key", "global_key0", "Key name")
	flag.Parse()
	fmt.Println("Outputs:", *filepath)
	fmt.Println("Appname:", *appName)
	fmt.Println("Scope:", *scope)
	fmt.Println("Key:", *key)
	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")
	err := secretStore.LoadSecrets(*appName, *scope)
	if err != nil {
		panic(err)
	} 

	val, err := secretStore.GetSecret(*appName, *key, *scope)
	if err != nil {
		panic(err)
	}
	fmt.Println("Value:", val)

	apps, minues, err := secretStore.ListAppNames()
	if err != nil {
		panic(err)
	}
	for app := range apps {
		fmt.Println("App name:", app)
	}
	for mine := range minues {
		fmt.Println("Mine name:", mine)
	}
}
