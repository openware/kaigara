package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// App contains a map of scopes(public, private, secret) with secrets to be loaded
type App struct {
	Scopes map[string]map[string]interface{} `yaml:"scopes"`
}

// SecretsFile contains secrets a map of Apps containing secrets to be loaded into the SecretStore
type SecretsFile struct {
	Secrets map[string]App `yaml:"secrets"`
}

func saveCmd() error {
	ss, err := loadStorageService()
	if err != nil {
		return fmt.Errorf("storage service init failed: %s", err)
	}

	data, err := os.ReadFile(SecretsPath)
	if err != nil {
		return err
	}

	secrets := SecretsFile{}
	if err := yaml.Unmarshal(data, &secrets); err != nil {
		return err
	}

	for app, scopes := range secrets.Secrets {
		for scope, data := range scopes.Scopes {
			if err := ss.Read(app, scope); err != nil {
				return err
			}

			delete(data, "version")
			for k, v := range data {
				log.Printf("INF: setting %s.%s.%s", app, scope, k)
				if err := ss.SetEntry(app, scope, k, v); err != nil {
					return err
				}
			}

			if err = ss.Write(app, scope); err != nil {
				return err
			}
		}
	}

	return nil
}
