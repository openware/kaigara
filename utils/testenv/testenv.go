package testenv

import (
	"os"
	"path"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
	"gopkg.in/yaml.v3"
)

var envs = map[string]map[string]map[string]string{}

func setupStore(ss types.Storage) {
	for name, app := range envs {
		for scope, elem := range app {
			err := ss.Read(name, scope)
			if err != nil {
				panic(err)
			}

			secrets, err := ss.GetEntries(name, scope)
			if err != nil {
				panic(err)
			}

			isSave := false
			for key, val := range elem {
				if _, ok := secrets[key]; !ok {
					isSave = true
					err = ss.SetEntry(name, scope, key, val)
					if err != nil {
						panic(err)
					}
				}
			}
			if isSave {
				err = ss.Write(name, scope)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func GetTestStorage(relativePath string, conf *config.KaigaraConfig) types.Storage {
	ss, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	workdirPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	absolutePath := path.Join(workdirPath, relativePath)
	envFile, err := os.ReadFile(absolutePath)
	if err != nil {
		panic(err)
	}

	envs = make(map[string]map[string]map[string]string)
	err = yaml.Unmarshal(envFile, &envs)
	if err != nil {
		panic(err)
	}

	setupStore(ss)

	return ss
}
