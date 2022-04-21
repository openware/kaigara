package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"

	"github.com/openware/kaigara/cmd/testenv"
	"github.com/openware/kaigara/pkg/config"
)

var deploymentID string
var sqlConfig database.Config

func TestMain(m *testing.M) {
	deploymentID = "opendax_uat"
	sqlConfig = database.Config{
		Driver: "mysql",
		Host:   os.Getenv("DATABASE_HOST"),
		Port:   os.Getenv("DATABASE_PORT"),
		User:   "root",
		Name:   "kaigara_" + deploymentID,
		Pass:   "",
		Pool:   1,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestKaidumpListAppNames(t *testing.T) {
	conf = &config.KaigaraConfig{
		Storage:       "sql",
		VaultAddr:     os.Getenv("KAIGARA_VAULT_ADDR"),
		VaultToken:    os.Getenv("KAIGARA_VAULT_TOKEN"),
		DeploymentID:  deploymentID,
		Scopes:        "private,secret",
		AppNames:      "finex,frontdex,gotrue,postgrest,realtime,storage",
		EncryptMethod: "aes",
		AesKey:        "changemechangemechangeme",
		DBConfig:      sqlConfig,
	}
	ss := testenv.GetStorage(conf)
	b := kaidumpRun(ss)
	assert.NotNil(t, b)
	fmt.Print(b.String())
}
