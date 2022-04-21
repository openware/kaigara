package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"

	"strings"

	"github.com/openware/pkg/ika"
	"gopkg.in/yaml.v3"
)

var conf = &config.KaigaraConfig{}

func main() {
	// Parse flags
	filepath := flag.String("f", "outputs.yaml", "Outputs file path to save secrets from vault")
	flag.Parse()

	// Initialize and write to Vault stores for every component
	if err := ika.ReadConfig("", conf); err != nil {
		panic(err)
	}

	secretStore, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	b := kaidumpRun(secretStore)
	fmt.Println(b.String())

	// Write secrets into filepath
	err = ioutil.WriteFile(*filepath, b.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	log.Printf("INF: dump saved into %s\n", *filepath)
}

func kaidumpRun(ss types.Storage) bytes.Buffer {
	// Get the list of App names from vault
	apps, err := ss.ListAppNames()
	if err != nil {
		panic(err)
	}

	// Get the list of scopes by Splitting KAIGARA_SCOPES env
	scopesList := strings.Split(conf.Scopes, ",")
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
			err := ss.Read(app, scope)
			if err != nil {
				panic(err)
			}
			secrets, err := ss.GetEntries(app, scope)
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
	err = yamlEncoder.Encode(&secretsMap)
	if err != nil {
		panic(err)
	}

	return b
}
