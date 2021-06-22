package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppNamesToLoggingName(t *testing.T) {
	cnf.AppNames = "peatio,peatio_daemons"
	assert.Equal(t, "peatio&peatio_daemons", appNamesToLoggingName())

	cnf.AppNames = "peatio"
	assert.Equal(t, "peatio", appNamesToLoggingName())
	assert.NotEqual(t, "peatio&", appNamesToLoggingName())
	assert.NotEqual(t, "&peatio", appNamesToLoggingName())
}
