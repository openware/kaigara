package config

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// type MockedConfig struct {
// 	config map[string]interface{}
// }

// func (m *MockedConfig) ListEntries() map[string]interface{} {
// 	return m.config
// }

// func TestBuildCmdEnvFilters(t *testing.T) {
// 	cnf := &MockedConfig{
// 		config: map[string]interface{}{
// 			"my_var":          "my_value",
// 			"AN_OTHER_SECRET": "secret",
// 		},
// 	}
// 	env := []string{
// 		"ANYTHING=must_be_kept",
// 		"KAIGARA_ANYTHING=must_be_ignored",
// 	}
// 	r := BuildCmdEnv(cnf, env)
// 	assert.Equal(t, map[string]*File{}, r.Files)
// 	assert.ElementsMatch(t, []string{
// 		"ANYTHING=must_be_kept",
// 		"AN_OTHER_SECRET=secret",
// 		"MY_VAR=my_value",
// 	}, r.Vars)
// }

// func TestBuildCmdEnvFileUpperCase(t *testing.T) {
// 	cnf := &MockedConfig{
// 		config: map[string]interface{}{
// 			"ANYTHING":           "must_be_set",
// 			"KFILE_NAME_PATH":    "config/config.json",
// 			"KFILE_NAME_CONTENT": `{"app":"example"}`,
// 		},
// 	}
// 	env := []string{}
// 	assert.Equal(t, &Env{
// 		Vars: []string{
// 			"ANYTHING=must_be_set",
// 		},
// 		Files: map[string]*File{
// 			"NAME": {
// 				Path:    "config/config.json",
// 				Content: `{"app":"example"}`,
// 			},
// 		},
// 	}, BuildCmdEnv(cnf, env))
// }

// func TestBuildCmdEnvFileLowerCase(t *testing.T) {
// 	cnf := &MockedConfig{
// 		config: map[string]interface{}{
// 			"anything":           "must_be_set",
// 			"kfile_name_path":    "config/config.json",
// 			"kfile_name_content": `{"app":"example"}`,
// 		},
// 	}
// 	env := []string{}
// 	assert.Equal(t, &Env{
// 		Vars: []string{
// 			"ANYTHING=must_be_set",
// 		},
// 		Files: map[string]*File{
// 			"NAME": {
// 				Path:    "config/config.json",
// 				Content: `{"app":"example"}`,
// 			},
// 		},
// 	}, BuildCmdEnv(cnf, env))
// }
