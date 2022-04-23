package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openware/kaigara/cmd/testenv"
	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
)

func TestMain(m *testing.M) {
	var err error
	if conf, err = config.NewKaigaraConfig(); err != nil {
		panic(err)
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestKaidumpListAppNames(t *testing.T) {
	ss := testenv.GetStorage(conf)

	b := kaidumpRun(ss)
	assert.NotNil(t, b)

	appNames := strings.Split(conf.AppNames, ",")
	scopes := strings.Split(conf.Scopes, ",")
	if err := storage.CleanAll(ss, appNames, scopes); err != nil {
		panic(err)
	}
}
