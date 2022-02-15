package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/kaigara/cmd/env"
	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage/sql"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var deploymentID string
var sqlConfig *database.Config

func TestMain(m *testing.M) {
	deploymentID = "opendax_uat"
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

	// Cleanup data
	sqlDB, err := database.Connect(sqlConfig)
	if err != nil {
		panic(err)
	}
	tx := sqlDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&sql.Data{})
	if tx.Error != nil {
		panic(tx.Error)
	}

	os.Exit(code)
}

func TestKaidumpListAppNames(t *testing.T) {
	cnf = &config.KaigaraConfig{
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
	store := env.GetStorage(cnf)
	b := kaidumpRun(store)
	assert.NotNil(t, b)
	fmt.Print(b.String())
}
