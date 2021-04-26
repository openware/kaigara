package main

import (
	"flag"
	"fmt"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/vault"

	"github.com/openware/pkg/ika"
	"strings"
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
	keyName := flag.String("k", "key1", "key name")
	flag.Parse()

	// Initialize and write to Vault stores for every component
	initConfig()
	secretStore := getVaultService("global")

	// Get the list of scopes by Splitting KAIGARA_SCOPES env
	scopesList := strings.Split(cnf.Scopes, ",")
	if len(scopesList) <= 0 {
		panic("Wrong KAIGARA_SCOPES")
	}

	for _, scope := range scopesList {
		err := secretStore.LoadSecrets(cnf.AppName, scope)
		if err != nil {
			panic(err)
		}

		err = secretStore.DeleteSecret(cnf.AppName, *keyName, scope)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Success")
}
