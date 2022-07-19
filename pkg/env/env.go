package env

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
)

var kfile = regexp.MustCompile("(?i)^KFILE_(.*)_(PATH|CONTENT)$")

func GetCompositeValueB64(v interface{}) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

// BuildCmdEnv reads secrets from all secretStores and scopes passed to it and loads them into an Env and returns a *Env
func BuildCmdEnv(appNames []string, ss types.Storage, currentEnv, scopes []string) (*config.Env, error) {
	env := &config.Env{
		Vars:  []string{},
		Files: map[string]*config.File{},
	}

	for _, v := range currentEnv {
		if !strings.HasPrefix(v, "KAIGARA_") {
			env.Vars = append(env.Vars, v)
		}
	}

	for _, appName := range append([]string{"global"}, appNames...) {
		for _, scope := range scopes {
			if err := ss.Read(appName, scope); err != nil {
				return nil, err
			}

			secrets, err := ss.GetEntries(appName, scope)
			if err != nil {
				return nil, err
			}

			for k, v := range secrets {
				var val string

				_, isMap := v.(map[string]interface{})
				_, isArray := v.([]interface{})
				if isMap || isArray {
					if compVal, err := GetCompositeValueB64(v); err != nil {
						return nil, err
					} else {
						val = compVal
					}
				}

				// Handle bool and json.Number
				if tmp, ok := v.(bool); ok {
					val = strconv.FormatBool(tmp)
				}

				if tmp, ok := v.(json.Number); ok {
					val = string(tmp)
				}

				if tmp, ok := v.(float64); ok {
					val = strconv.FormatFloat(tmp, 'f', -1, 64)
				}

				// Skip if the var can't be asserted to string
				if val == "" {
					if tmp, ok := v.(string); ok {
						val = tmp
					} else {
						continue
					}
				}

				m := kfile.FindStringSubmatch(k)

				if m == nil {
					env.Vars = append(env.Vars, strings.ToUpper(k)+"="+val)
					continue
				}
				name := strings.ToUpper(m[1])
				suffix := strings.ToUpper(m[2])

				f, ok := env.Files[name]
				if !ok {
					f = &config.File{}
					env.Files[name] = f
				}
				switch suffix {
				case "PATH":
					f.Path = v.(string)
				case "CONTENT":
					f.Content = v.(string)
				default:
					log.Printf("ERR: unexpected prefix in config key: %s\n", k)
				}
			}
		}
	}

	return env, nil
}
