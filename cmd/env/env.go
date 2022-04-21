package env

import (
	"io/ioutil"
	"os"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
	"gopkg.in/yaml.v3"
)

var envs = map[string]map[string]map[string]string{}

func setupStore(store types.Storage) {
	for name, app := range envs {
		for scope, elem := range app {
			err := store.Read(name, scope)
			if err != nil {
				panic(err)
			}

			secrets, err := store.GetEntries(name, scope)
			if err != nil {
				panic(err)
			}

			isSave := false
			for key, val := range elem {
				if _, ok := secrets[key]; !ok {
					isSave = true
					err = store.SetEntry(name, scope, key, val)
					if err != nil {
						panic(err)
					}
				}
			}
			if isSave {
				err = store.Write(name, scope)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func GetStorage(conf *config.KaigaraConfig) types.Storage {
	ss, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	envFile, err := ioutil.ReadFile(path + "/../env/env.yml")
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
