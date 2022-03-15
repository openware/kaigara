package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
	"github.com/stretchr/testify/assert"
)

type mockSecretStorage struct {
	types.Storage
}

func (mockSS *mockSecretStorage) Read(_, _ string) error { return nil }

func (mockSS *mockSecretStorage) GetEntries(appName, scope string) (map[string]interface{}, error) {
	app, ok := testStore[appName]
	if !ok {
		return nil, fmt.Errorf("no such app: %s", appName)
	}

	entries, ok := app[scope]
	if !ok {
		return nil, fmt.Errorf("no such scope for app %s: %s", appName, scope)
	}

	return entries, nil
}

var testStore = map[string]map[string]map[string]interface{}{
	"finex": {
		"private": map[string]interface{}{
			"finex_influx_host": "influxdb-0.core,influxdb-1.core,influxdb-2.core",
			"finex_log_level":   "debug",
			"finex_mode":        "prod",
			"finex_vault_addr":  "http://vault.core:8200",
		},
		"public": map[string]interface{}{},
		"secret": map[string]interface{}{
			"finex_license_key":              "eiJohdo9eish3Cooshus",
			"finex_vault_broker_private_key": "ujahDoo1ohmie7taipox",
			"finex_vault_rpc_url":            "wss://rinkeby.infura.io/ws/v3/Uzeep6eiGoozui7ohsh9",
		},
	},
	"frontdex": {
		"private": map[string]interface{}{
			"next_public_chain_id":         "4",
			"next_public_custody_contract": "Feibie7ooCachie3eePh",
		},
		"public": map[string]interface{}{},
		"secret": map[string]interface{}{
			"next_public_infura_id": "bie9niNgoohadoorai5a",
		},
	},
	"global": {
		"private": map[string]interface{}{},
		"public":  map[string]interface{}{},
		"secret": map[string]interface{}{
			"database_host": "0.0.0.0",
			"database_port": "3306",
			"postgres_host": "0.0.0.0",
			"postgres_port": "5432",
		},
	},
}

func Test_printExactEnv(t *testing.T) {
	mockSS := &mockSecretStorage{}
	testConf := &config.KaigaraConfig{
		Scopes:   "public,private,secret",
		AppNames: "finex,frontdex,global",
	}

	for i := 0; i < 5; i++ {
		for _, app := range testStore {
			for _, entries := range app {
				for key, value := range entries {
					envVariable := strings.ToUpper(key)
					envValueExpected := value

					t.Run(fmt.Sprintf("Test print %s %d", envVariable, i), func(t *testing.T) {
						envValueActual, err := getExactEnv(testConf, mockSS, envVariable)
						assert.Equal(t, envValueExpected, envValueActual)
						assert.NoError(t, err)
					})
				}
			}
		}
	}
}
