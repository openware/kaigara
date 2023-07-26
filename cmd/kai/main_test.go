package main

import (
	"os"
	"testing"

	"github.com/openware/kaigara/pkg/config"
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
