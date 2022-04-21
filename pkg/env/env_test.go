package env

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
)

var scopes []string = []string{"secret"}
var vaultAddr string = os.Getenv("KAIGARA_VAULT_ADDR")
var vaultToken string = os.Getenv("KAIGARA_VAULT_TOKEN")
var deploymentID string = "kaigara_test"
var ss, _ = storage.NewVaultService(vaultAddr, vaultToken, deploymentID)

func TestBuildCmdEnvFromSecretss(t *testing.T) {
	appName := "test1"
	appNames := []string{"test1"}

	env := []string{
		"ANYTHING=must_be_kept",
		"KAIGARA_ANYTHING=must_be_ignored",
	}

	err := ss.Read(appName, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "key_"+scopes[0], "value_"+scopes[0])
	assert.NoError(t, err)

	err = ss.Write(appName, scopes[0])
	assert.NoError(t, err)

	err = ss.Read("global", "secret")
	assert.NoError(t, err)

	err = ss.SetEntry("global", scopes[0], "key_global", "value_global")
	assert.NoError(t, err)

	err = ss.Write("global", scopes[0])
	assert.NoError(t, err)

	r, err := BuildCmdEnv(appNames, ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, map[string]*config.File{}, r.Files)
	assert.ElementsMatch(t, []string{
		"ANYTHING=must_be_kept",
		"KEY_SECRET=value_secret",
		"KEY_GLOBAL=value_global",
	}, r.Vars)
}

func TestLoadNumberAndBool(t *testing.T) {
	appName := "test2"
	appNames := []string{"test2"}

	scopes = []string{"public"}
	env := []string{}

	err := ss.Read(appName, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "key_number", json.Number("1337"))
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "key_bool", true)
	assert.NoError(t, err)

	err = ss.Write(appName, scopes[0])
	assert.NoError(t, err)

	r, err := BuildCmdEnv(appNames, ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, map[string]*config.File{}, r.Files)
	assert.ElementsMatch(t, []string{
		"KEY_NUMBER=1337",
		"KEY_BOOL=true",
	}, r.Vars)
}

func TestBuildCmdEnvFileUpperCase(t *testing.T) {
	appName := "test3"
	appNames := []string{"test3"}

	err := ss.Read(appName, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "ANYTHING", "must_be_set")
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "KFILE_NAME_PATH", "config/config.json")
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "KFILE_NAME_CONTENT", `{"app":"example"}`)
	assert.NoError(t, err)

	err = ss.Write(appName, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames, ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &config.Env{
		Vars: []string{
			"ANYTHING=must_be_set",
		},
		Files: map[string]*config.File{
			"NAME": {
				Path:    "config/config.json",
				Content: `{"app":"example"}`,
			},
		},
	}, r)
}

func TestBuildCmdEnvFileLowerCase(t *testing.T) {
	appName := "test4"
	appNames := []string{"test4"}

	err := ss.Read(appName, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "anything", "must_be_set")
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "kfile_name_path", "config/config.json")
	assert.NoError(t, err)

	err = ss.SetEntry(appName, scopes[0], "kfile_name_content", `{"app":"example"}`)
	assert.NoError(t, err)

	err = ss.Write(appName, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames, ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &config.Env{
		Vars: []string{
			"ANYTHING=must_be_set",
		},
		Files: map[string]*config.File{
			"NAME": {
				Path:    "config/config.json",
				Content: `{"app":"example"}`,
			},
		},
	}, r)
}

func TestBuildCmdEnvSeveralAppNames(t *testing.T) {
	appNameFirst := "test5"
	appNameSecond := "test6"
	appNames := []string{"test5", "test6"}

	err := ss.Read(appNameFirst, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appNameFirst, scopes[0], "anything_5", "must_be_set")
	assert.NoError(t, err)

	err = ss.Write(appNameFirst, scopes[0])
	assert.NoError(t, err)

	err = ss.Read(appNameSecond, scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appNameSecond, scopes[0], "anything_6", "must_be_set")
	assert.NoError(t, err)

	err = ss.Write(appNameSecond, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames, ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &config.Env{
		Vars: []string{
			"ANYTHING_5=must_be_set",
			"ANYTHING_6=must_be_set",
		},
		Files: map[string]*config.File{},
	}, r)
}
