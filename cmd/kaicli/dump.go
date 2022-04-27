package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/openware/kaigara/types"
)

func dumpCmd() error {
	b := kaidumpRun(ss)
	fmt.Println(b.String())

	if err := ioutil.WriteFile(SecretsPath, b.Bytes(), 0644); err != nil {
		return err
	}

	log.Printf("INF: dump saved into %s\n", SecretsPath)
	return nil
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
			if err := ss.Read(app, scope); err != nil {
				panic(err)
			}

			secrets, err := ss.GetEntries(app, scope)
			if err != nil {
				panic(err)
			}

			delete(secrets, "version")
			if len(secrets) > 0 {
				scopeMap[scope] = secrets
			}
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
