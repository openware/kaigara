package main

import (
	"bytes"
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

func (mockSS *mockSecretStorage) Read(_, _ string) error {
	return nil
}

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
			"FINEX_INFLUX_HOST": "influxdb-0.core,influxdb-1.core,influxdb-2.core",
			"FINEX_LOG_LEVEL":   "debug",
			"FINEX_MODE":        "prod",
			"FINEX_VAULT_ADDR":  "http://vault.core:8200",
			"FINEX_STRUCT_ENV": []interface{}{
				map[string]interface{}{
					"field1": "value2",
				},
				map[string]interface{}{
					"field3": "",
				},
			},
		},
		"public": map[string]interface{}{},
		"secret": map[string]interface{}{
			"FINEX_LICENSE_KEY":              "eiJohdo9eish3Cooshus",
			"FINEX_VAULT_BROKER_PRIVATE_KEY": "ujahDoo1ohmie7taipox",
			"FINEX_VAULT_RPC_URL":            "wss://rinkeby.infura.io/ws/v3/Uzeep6eiGoozui7ohsh9",
		},
	},
	"frontdex": {
		"private": map[string]interface{}{
			"NEXT_PUBLIC_CHAIN_ID":         "4",
			"NEXT_PUBLIC_CUSTODY_CONTRACT": "Feibie7ooCachie3eePh",
		},
		"public": map[string]interface{}{},
		"secret": map[string]interface{}{
			"NEXT_PUBLIC_INFURA_ID": "bie9niNgoohadoorai5a",
		},
	},
	"global": {
		"private": map[string]interface{}{},
		"public":  map[string]interface{}{},
		"secret": map[string]interface{}{
			"DATABASE_HOST": "0.0.0.0",
			"DATABASE_PORT": "3306",
			"POSTGRES_HOST": "0.0.0.0",
			"POSTGRES_PORT": "5432",
		},
	},
}

func TestKaienvRun(t *testing.T) {
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
					envValueExpected, err := envValueToString(value)
					assert.NoError(t, err)

					t.Run(fmt.Sprintf("Test print %s %d", envVariable, i), func(t *testing.T) {
						var buff bytes.Buffer
						err := kaienvRun(testConf, mockSS, []string{envVariable}, &buff)
						envValueActual := buff.String()
						assert.Equal(t, envValueExpected, envValueActual)
						assert.NoError(t, err)
					})
				}
			}
		}
	}
}
