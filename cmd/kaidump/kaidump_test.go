package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/kaigara/cmd/env"
	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
)

var deploymentID string
var appNames []string
var sqlConfig *database.Config

func TestMain(m *testing.M) {
	deploymentID = "opendax_uat"
	appNames = []string{"finex", "frontdex", "gotrue", "postgrest", "realtime", "storage"}
	sqlConfig = &database.Config{
		Driver: "mysql",
		Host:   os.Getenv("DATABASE_HOST"),
		Port:   os.Getenv("DATABASE_PORT"),
		Name:   "kaigara_dev",
		User:   "root",
		Pass:   "",
		Pool:   1,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestKaidumpListAppNames(t *testing.T) {
	db := sqlConfig
	cfg := &config.KaigaraConfig{
		SecretStore:  "sql",
		VaultAddr:    os.Getenv("KAIGARA_VAULT_ADDR"),
		VaultToken:   os.Getenv("KAIGARA_VAULT_TOKEN"),
		DeploymentID: deploymentID,
		Scopes:       "private,secret",
		AppNames:     "finex,frontdex,gotrue,postgrest,realtime,storage",
	}

	cnf = cfg
	sqlCnf = db
	store := env.GetStorage(cfg, db)
	b := kaidumpRun(store)
	assert.NotNil(t, b)
	fmt.Print(b.String())
}
