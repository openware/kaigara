package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/openware/kaigara/types"
)

func dumpCmd() error {
	ss, err := loadStorageService()
	if err != nil {
		return fmt.Errorf("storage service init failed: %s", err)
	}

	b := kaidumpRun(ss)
	fmt.Println(b.String())

	if err := os.WriteFile(SecretsPath, b.Bytes(), 0644); err != nil {
		return err
	}

	log.Printf("INF: dump saved into %s\n", SecretsPath)
	return nil
}

func kaidumpRun(ss types.Storage) bytes.Buffer {
	var (
		apps []string
		err  error
	)

	if conf.AppNames == "" {
		apps, err = ss.ListAppNames()
		if err != nil {
			panic(err)
		}
	} else {
		apps = strings.Split(conf.AppNames, ",")
	}

	// Get the list of scopes by Splitting KAIGARA_SCOPES env
	scopesList := strings.Split(conf.Scopes, ",")
	if len(scopesList) <= 0 {
		panic("Please specify KAIGARA_SCOPES env var")
	} else if conf.Storage == "k8s" {
		scopesList = []string{"secret"}
	}

	// Create Secrets map
	secretsMap := make(map[string]map[string]map[string]interface{})

	// Get the secrets from vault
	for _, app := range apps {
		appMap := make(map[string]map[string]interface{})
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
				appMap[scope] = secrets
			}
		}
		secretsMap[app] = appMap
	}

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
