package main

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/logstream"
	"github.com/openware/kaigara/pkg/sql"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/utils/testenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

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

var testdataPath = "../testdata/testenv.yml"

func TestMain(m *testing.M) {
	var err error
	if conf, err = config.NewKaigaraConfig(); err != nil {
		panic(err)
	}

	ls, err = logstream.NewRedisClient(conf.RedisURL)
	if err != nil {
		log.Printf("WRN: %s", err.Error())
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestAppNamesToLoggingName(t *testing.T) {
	conf.AppNames = "peatio,peatio_daemons"
	assert.Equal(t, "peatio&peatio_daemons", appNamesToLoggingName())

	conf.AppNames = "peatio"
	assert.Equal(t, "peatio", appNamesToLoggingName())
	assert.NotEqual(t, "peatio&", appNamesToLoggingName())
	assert.NotEqual(t, "&peatio", appNamesToLoggingName())
}

func TestKaigaraPrintenvVault(t *testing.T) {
	conf.Storage = "vault"
	conf.AppNames = "finex,frontdex,gotrue,postgrest,realtime,storage"
	ss := testenv.GetTestStorage(testdataPath, conf)

	for _, v := range vars {
		kaigaraRun(ss, "printenv", []string{v})
	}

	appNames := strings.Split(conf.AppNames, ",")
	scopes := strings.Split(conf.Scopes, ",")
	if err := storage.CleanAll(ss, appNames, scopes); err != nil {
		panic(err)
	}
}

func TestKaigaraPrintenvSql(t *testing.T) {
	conf.Storage = "sql"
	conf.AppNames = "finex,frontdex,gotrue,postgrest,realtime,storage"
	ss := testenv.GetTestStorage(testdataPath, conf)

	for _, v := range vars {
		kaigaraRun(ss, "printenv", []string{v})
	}

	// Cleanup data
	sqlDB, err := sql.Connect(&conf.DBConfig)
	if err != nil {
		t.Fatal(err)
	}

	tx := sqlDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&sql.Data{})
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
}
