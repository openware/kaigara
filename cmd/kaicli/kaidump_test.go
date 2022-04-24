package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openware/kaigara/cmd/testenv"
	"github.com/openware/kaigara/pkg/storage"
)

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
