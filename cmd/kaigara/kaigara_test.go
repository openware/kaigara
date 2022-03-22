package main

import (
	"os"
	"testing"

	"github.com/openware/kaigara/cmd/env"
	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage/sql"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var deploymentID = "opendax_uat"
var sqlCnf = database.Config{
	Driver: "mysql",
	Host:   os.Getenv("DATABASE_HOST"),
	Port:   os.Getenv("DATABASE_PORT"),
	Name:   "kaigara_dev",
	User:   "root",
	Pass:   "",
	Pool:   1,
}

var vars = []string{
	"FINEX_DATABASE_USERNAME",
	"FINEX_DATABASE_PASSWORD",
	"FINEX_DATABASE_NAME",
	"FINEX_DATABASE_HOST",
	"FINEX_INFLUX_USERNAME",
	"FINEX_INFLUX_PASSWORD",
	"FINEX_INFLUX_HOST",
	"GOTRUE_DATABASE_USERNAME",
	"GOTRUE_DATABASE_PASSWORD",
	"GOTRUE_DATABASE_NAME",
	"GOTRUE_DATABASE_HOST",
	"PGRST_DB_USERNAME",
	"PGRST_DB_PASS",
	"PGRST_DB_NAME",
	"PGRST_DB_HOST",
	"REALTIME_DB_USERNAME",
	"REALTIME_DB_PASS",
	"REALTIME_DB_NAME",
	"REALTIME_DB_HOST",
}

func TestMain(m *testing.M) {
	cnf = &config.KaigaraConfig{
		VaultAddr:     os.Getenv("KAIGARA_VAULT_ADDR"),
		VaultToken:    os.Getenv("KAIGARA_VAULT_TOKEN"),
		DeploymentID:  deploymentID,
		Scopes:        "public,private,secret",
		AppNames:      "finex,frontdex,gotrue,postgrest,realtime,storage",
		EncryptMethod: "transit",
		DBConfig:      sqlCnf,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestAppNamesToLoggingName(t *testing.T) {
	cnf.AppNames = "peatio,peatio_daemons"
	assert.Equal(t, "peatio&peatio_daemons", appNamesToLoggingName())

	cnf.AppNames = "peatio"
	assert.Equal(t, "peatio", appNamesToLoggingName())
	assert.NotEqual(t, "peatio&", appNamesToLoggingName())
	assert.NotEqual(t, "&peatio", appNamesToLoggingName())
}

func TestKaigaraPrintenvVault(t *testing.T) {
	cnf.Storage = "vault"
	cnf.AppNames = "finex,frontdex,gotrue,postgrest,realtime,storage"
	store := env.GetStorage(cnf)
	ls := initLogStream()

	for _, v := range vars {
		kaigaraRun(ls, store, "printenv", []string{v})
	}
}

func TestKaigaraPrintenvSql(t *testing.T) {
	cnf.Storage = "sql"
	cnf.AppNames = "finex,frontdex,gotrue,postgrest,realtime,storage"
	store := env.GetStorage(cnf)
	ls := initLogStream()

	for _, v := range vars {
		kaigaraRun(ls, store, "printenv", []string{v})
	}

	// Cleanup data
	sqlDB, err := database.Connect(&sqlCnf)
	if err != nil {
		panic(err)
	}
	tx := sqlDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&sql.Data{})
	if tx.Error != nil {
		panic(tx.Error)
	}
}
