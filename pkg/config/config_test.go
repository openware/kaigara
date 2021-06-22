package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/kaigara/types"
	"github.com/stretchr/testify/assert"
)

var scopes []string = []string{"secret"}
var vaultAddr string = os.Getenv("KAIGARA_VAULT_ADDR")
var vaultToken string = os.Getenv("KAIGARA_VAULT_TOKEN")
var deploymentID string = "kaigara_test"
var secretStore types.SecretStore = vault.NewService(vaultAddr, vaultToken, deploymentID)

func TestBuildCmdEnvFromSecretStore(t *testing.T) {
	appName := "test1"
	appNames := []string{"test1"}

	env := []string{
		"ANYTHING=must_be_kept",
		"KAIGARA_ANYTHING=must_be_ignored",
	}

	err := secretStore.LoadSecrets(appName, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "key_"+scopes[0], "value_"+scopes[0], scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appName, scopes[0])
	assert.NoError(t, err)

	err = secretStore.LoadSecrets("global", "secret")
	assert.NoError(t, err)

	err = secretStore.SetSecret("global", "key_global", "value_global", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets("global", scopes[0])
	assert.NoError(t, err)

	r := BuildCmdEnv(appNames, secretStore, env, scopes)

	assert.Equal(t, map[string]*File{}, r.Files)
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

	err := secretStore.LoadSecrets(appName, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "key_number", json.Number("1337"), scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "key_bool", true, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appName, scopes[0])
	assert.NoError(t, err)

	r := BuildCmdEnv(appNames, secretStore, env, scopes)

	assert.Equal(t, map[string]*File{}, r.Files)
	assert.ElementsMatch(t, []string{
		"KEY_NUMBER=1337",
		"KEY_BOOL=true",
	}, r.Vars)
}

func TestBuildCmdEnvFileUpperCase(t *testing.T) {
	appName := "test3"
	appNames := []string{"test3"}

	err := secretStore.LoadSecrets(appName, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "ANYTHING", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "KFILE_NAME_PATH", "config/config.json", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "KFILE_NAME_CONTENT", `{"app":"example"}`, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appName, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	assert.Equal(t, &Env{
		Vars: []string{
			"ANYTHING=must_be_set",
		},
		Files: map[string]*File{
			"NAME": {
				Path:    "config/config.json",
				Content: `{"app":"example"}`,
			},
		},
	}, BuildCmdEnv(appNames, secretStore, env, scopes))
}

func TestBuildCmdEnvFileLowerCase(t *testing.T) {
	appName := "test4"
	appNames := []string{"test4"}

	err := secretStore.LoadSecrets(appName, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "anything", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "kfile_name_path", "config/config.json", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appName, "kfile_name_content", `{"app":"example"}`, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appName, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	assert.Equal(t, &Env{
		Vars: []string{
			"ANYTHING=must_be_set",
		},
		Files: map[string]*File{
			"NAME": {
				Path:    "config/config.json",
				Content: `{"app":"example"}`,
			},
		},
	}, BuildCmdEnv(appNames, secretStore, env, scopes))
}


func TestBuildCmdEnvSeveralAppNames(t *testing.T) {
	appNameFirst := "test5"
	appNameSecond := "test6"
	appNames := []string{"test5", "test6"}

	err := secretStore.LoadSecrets(appNameFirst, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appNameFirst, "anything_5", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appNameFirst, scopes[0])
	assert.NoError(t, err)

	err = secretStore.LoadSecrets(appNameSecond, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret(appNameSecond, "anything_6", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(appNameSecond, scopes[0])
	assert.NoError(t, err)

	env := []string{}
	assert.Equal(t, &Env{
		Vars: []string{
			"ANYTHING_5=must_be_set",
			"ANYTHING_6=must_be_set",
		},
		Files: map[string]*File{},
	}, BuildCmdEnv(appNames, secretStore, env, scopes))
}
