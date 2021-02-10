package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/kaigara/types"
	"github.com/stretchr/testify/assert"
)

var scopes []string = []string{"public"}
var vaultAddr string = os.Getenv("KAIGARA_VAULT_ADDR")
var vaultToken string = os.Getenv("KAIGARA_VAULT_TOKEN")
var deploymentID string = "kaigara_test"
var appName string = "peatio"
var secretStore types.SecretStore = vault.NewService(vaultAddr, vaultToken, appName, deploymentID)

func TestBuildCmdEnvFromSecretStore(t *testing.T) {
	secretStore.SetAppName("test1")

	env := []string{
		"ANYTHING=must_be_kept",
		"KAIGARA_ANYTHING=must_be_ignored",
	}

	err := secretStore.LoadSecrets(scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("key_"+scopes[0], "value_"+scopes[0], scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(scopes[0])
	assert.NoError(t, err)

	r := BuildCmdEnv([]types.SecretStore{secretStore}, env, scopes)

	fmt.Println(r.Vars)
	assert.Equal(t, map[string]*File{}, r.Files)
	assert.ElementsMatch(t, []string{
		"ANYTHING=must_be_kept",
		"KEY_PUBLIC=value_public",
	}, r.Vars)
}

func TestBuildCmdEnvFileUpperCase(t *testing.T) {
	secretStore.SetAppName("test2")

	err := secretStore.LoadSecrets(scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("ANYTHING", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("KFILE_NAME_PATH", "config/config.json", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("KFILE_NAME_CONTENT", `{"app":"example"}`, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(scopes[0])
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
	}, BuildCmdEnv([]types.SecretStore{secretStore}, env, scopes))
}

func TestBuildCmdEnvFileLowerCase(t *testing.T) {
	secretStore.SetAppName("test3")

	err := secretStore.LoadSecrets(scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("anything", "must_be_set", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("kfile_name_path", "config/config.json", scopes[0])
	assert.NoError(t, err)

	err = secretStore.SetSecret("kfile_name_content", `{"app":"example"}`, scopes[0])
	assert.NoError(t, err)

	err = secretStore.SaveSecrets(scopes[0])
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
	}, BuildCmdEnv([]types.SecretStore{secretStore}, env, scopes))
}
