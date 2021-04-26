package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/vault"

	"strings"

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

func main() {
	// Parse flags
	filepath := flag.String("a", "outputs.yaml", "Outputs file path to save secrets from vault")
	flag.Parse()

	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")

	// Get the list of App names from vault
	apps, err := secretStore.ListAppNames()
	if err != nil {
		panic(err)
	}

	// Get the list of scopes by Splitting KAIGARA_SCOPES env
	scopesList := strings.Split(cnf.Scopes, ",")
	if len(scopesList) <= 0 {
		panic("Wrong KAIGARA_SCOPES")
	}

	// Create Secrets map
	secretsMap := make(map[string]interface{})

	// Create App map
	appMap := make(map[string]interface{})

	// Get the secrets from vault
	for _, app := range apps {
		scopeMap := make(map[string]interface{})
		scopeInit := make(map[string]interface{})
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
		scopeInit["scopes"] = scopeMap
		appMap[app] = scopeInit
	}

	secretsMap["secrets"] = appMap

	// Encode into YAML
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&secretsMap)

	fmt.Print(b.String())

	// Write secrets into filepath
	err = ioutil.WriteFile(*filepath, b.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("# Saved the dump into %s\n", *filepath)
}
