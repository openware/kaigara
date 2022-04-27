package env

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/pkg/vault"
)

var appNames []string = []string{
	"global",
	"test1",
	"test2",
	"test3",
	"test4",
	"test5",
	"test6",
}
var scopes []string = []string{"secret", "public"}
var ss *vault.Service

func TestMain(m *testing.M) {
	conf, err := config.NewKaigaraConfig()
	if err != nil {
		panic(err)
	}

	ss, err = vault.NewService(conf.VaultAddr, conf.VaultToken, conf.DeploymentID)
	if err != nil {
		panic(err)
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	if err := storage.CleanAll(ss, appNames, scopes); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestBuildCmdEnvFromSecretss(t *testing.T) {
	env := []string{
		"ANYTHING=must_be_kept",
		"KAIGARA_ANYTHING=must_be_ignored",
	}

	err := ss.Read(appNames[1], scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[1], scopes[0], "key_"+scopes[0], "value_"+scopes[0])
	assert.NoError(t, err)

	err = ss.Write(appNames[1], scopes[0])
	assert.NoError(t, err)

	err = ss.Read(appNames[0], scopes[0])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[0], scopes[0], "key_global", "value_global")
	assert.NoError(t, err)

	err = ss.Write(appNames[0], scopes[0])
	assert.NoError(t, err)

	r, err := BuildCmdEnv(appNames[1:2], ss, env, scopes)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, map[string]*config.File{}, r.Files)
	assert.ElementsMatch(t, []string{
		"ANYTHING=must_be_kept",
		"KEY_SECRET=value_secret",
		"KEY_GLOBAL=value_global",
	}, r.Vars)

	//Cleanup global app
	entries, err := ss.ListEntries(appNames[0], scopes[0])
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if err := ss.DeleteEntry(appNames[0], scopes[0], entry); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadNumberAndBool(t *testing.T) {
	env := []string{}

	err := ss.Read(appNames[2], scopes[1])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[2], scopes[1], "key_number", json.Number("1337"))
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[2], scopes[1], "key_bool", true)
	assert.NoError(t, err)

	err = ss.Write(appNames[2], scopes[1])
	assert.NoError(t, err)

	r, err := BuildCmdEnv(appNames[2:3], ss, env, scopes)
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
	err := ss.Read(appNames[3], scopes[1])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[3], scopes[1], "ANYTHING", "must_be_set")
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[3], scopes[1], "KFILE_NAME_PATH", "config/config.json")
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[3], scopes[1], "KFILE_NAME_CONTENT", `{"app":"example"}`)
	assert.NoError(t, err)

	err = ss.Write(appNames[3], scopes[1])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames[3:4], ss, env, scopes)
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
	err := ss.Read(appNames[4], scopes[1])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[4], scopes[1], "anything", "must_be_set")
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[4], scopes[1], "kfile_name_path", "config/config.json")
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[4], scopes[1], "kfile_name_content", `{"app":"example"}`)
	assert.NoError(t, err)

	err = ss.Write(appNames[4], scopes[1])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames[4:5], ss, env, scopes)
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
	err := ss.Read(appNames[5], scopes[1])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[5], scopes[1], "anything_5", "must_be_set")
	assert.NoError(t, err)

	err = ss.Write(appNames[5], scopes[1])
	assert.NoError(t, err)

	err = ss.Read(appNames[6], scopes[1])
	assert.NoError(t, err)

	err = ss.SetEntry(appNames[6], scopes[1], "anything_6", "must_be_set")
	assert.NoError(t, err)

	err = ss.Write(appNames[6], scopes[1])
	assert.NoError(t, err)

	env := []string{}
	r, err := BuildCmdEnv(appNames[5:], ss, env, scopes)
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
