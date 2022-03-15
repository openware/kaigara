package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/database"
	"github.com/openware/pkg/ika"
)

func main() {
	conf := &config.KaigaraConfig{}
	if err := ika.ReadConfig("", conf); err != nil {
		panic(err)
	}

	db := &database.Config{}
	if err := ika.ReadConfig("", db); err != nil {
		panic(err)
	}
	conf.DBConfig = db

	ss, err := config.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		err = printAllEnvs(conf, ss)
	} else {
		var envValue interface{}
		envValue, err = getExactEnv(conf, ss, os.Args[1])
		fmt.Print(envValue)
	}

	if err != nil {
		panic(err)
	}
}

func getExactEnv(conf *config.KaigaraConfig, ss types.Storage, envVariable string) (interface{}, error) {
	for _, appName := range strings.Split(conf.AppNames, ",") {
		for _, scope := range strings.Split(conf.Scopes, ",") {
			if err := ss.Read(appName, scope); err != nil {
				return "", err
			}

			entries, err := ss.GetEntries(appName, scope)
			if err != nil {
				return "", err
			}

			if envValue, ok := entries[strings.ToLower(envVariable)]; ok {
				return envValue, nil
			}
		}
	}

	return "", fmt.Errorf("no value for such key: %s", envVariable)
}

func printAllEnvs(conf *config.KaigaraConfig, ss types.Storage) error {
	for _, appName := range strings.Split(conf.AppNames, ",") {
		for _, scope := range strings.Split(conf.Scopes, ",") {
			if err := ss.Read(appName, scope); err != nil {
				return err
			}

			entries, err := ss.GetEntries(appName, scope)
			if err != nil {
				return err
			}

			for secret, value := range entries {
				envVariable := strings.ToUpper(secret)
				envValue, ok := value.(string)
				if !ok {
					return fmt.Errorf("value of secret %s is not string: %+v", envVariable, value)
				}

				fmt.Printf("%s=%s\n", envVariable, envValue)
			}
		}
	}

	return nil
}
