package env

import (
	"io/ioutil"
	"os"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/database"
	"gopkg.in/yaml.v3"
)

var envs = map[string]map[string]map[string]string{}

func setupStore(store types.Storage) {
	for name, app := range envs {
		for scope, elem := range app {
			store.Read(name, scope)
			secrets, err := store.GetEntries(name, scope)
			if err != nil {
				panic(err)
			}
			isSave := false
			for key, val := range elem {
				if _, ok := secrets[key]; !ok {
					isSave = true
					store.SetEntry(name, scope, key, val)
				}
			}
			if isSave {
				store.Write(name, scope)
			}
		}
	}
}

func GetStorage(cfg *config.KaigaraConfig, db *database.Config) types.Storage {
	store, err := config.GetStorageService(cfg, db)
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
	setupStore(store)

	return store
}
