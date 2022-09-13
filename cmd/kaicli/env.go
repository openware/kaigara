package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/env"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/kli"
)

func envCmd(cmd *kli.Command) func() error {
	return func() error {
		return kaienvRun(conf, ss, cmd.OtherArgs(), os.Stdout)
	}
}

func kaienvRun(conf *config.KaigaraConfig, ss types.Storage, params []string, out io.Writer) error {
	env, err := readAllEnv(conf, ss)
	if err != nil {
		return err
	}

	if len(params) < 1 {
		for envVariable, envValue := range env {
			if compVal, err := envValueToString(envValue); err != nil {
				return err
			} else {
				fmt.Fprintf(out, "%s=%s\n", envVariable, compVal)
			}
		}
	} else {
		envVariable := params[0]
		if envValue, ok := env[envVariable]; !ok {
			return fmt.Errorf("no value for such key: %s", envVariable)
		} else if compVal, err := envValueToString(envValue); err != nil {
			return err
		} else {
			fmt.Fprint(out, compVal)
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

			delete(entries, "version")
			for envVariable, envValue := range entries {
				env[envVariable] = envValue
			}
		}
	}

	return env, nil
}

func envValueToString(value interface{}) (string, error) {
	if envValue, ok := value.(string); ok {
		return fmt.Sprintf("\"%s\"", envValue), nil
	}

	_, isMap := value.(map[string]interface{})
	_, isArray := value.([]interface{})
	if isMap || isArray {
		if compVal, err := env.GetCompositeValueB64(value); err != nil {
			return "", err
		} else {
			return compVal, nil
		}
	}

	return fmt.Sprintf("%v", value), nil
}
