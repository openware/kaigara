package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/ika"
)

func main() {
	conf := &config.KaigaraConfig{}
	if err := ika.ReadConfig("", conf); err != nil {
		panic(err)
	}

	ss, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	if err := kaienvRun(conf, ss, os.Args, os.Stdout); err != nil {
		panic(err)
	}
}

func kaienvRun(conf *config.KaigaraConfig, ss types.Storage, params []string, out io.Writer) error {
	env, err := readAllEnv(conf, ss)
	if err != nil {
		return err
	}

	if len(params) < 2 {
		for envVariable, envValue := range env {
			fmt.Fprintf(out, "%s=%s\n", strings.ToUpper(envVariable), envValueToString(envValue))
		}
	} else {
		envVariable := params[1]
		if envValue, ok := env[strings.ToLower(envVariable)]; ok {
			fmt.Fprint(out, envValue)
		} else {
			return fmt.Errorf("no value for such key: %s", envVariable)
		}
	}

	return nil
}

func readAllEnv(conf *config.KaigaraConfig, ss types.Storage) (map[string]interface{}, error) {
	env := make(map[string]interface{})

	for _, appName := range strings.Split(conf.AppNames, ",") {
		for _, scope := range strings.Split(conf.Scopes, ",") {
			if err := ss.Read(appName, scope); err != nil {
				return nil, err
			}

			entries, err := ss.GetEntries(appName, scope)
			if err != nil {
				return nil, err
			}

			for envVariable, envValue := range entries {
				env[envVariable] = envValue
			}
		}
	}

	return env, nil
}

func envValueToString(value interface{}) string {
	if envValue, ok := value.(string); ok {
		return fmt.Sprintf("\"%s\"", envValue)
	}

	return fmt.Sprintf("%v", value)
}
