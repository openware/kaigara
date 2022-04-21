package storage

import (
	"fmt"
	"log"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/encryptor"
	"github.com/openware/kaigara/types"
)

func GetStorageService(conf *config.KaigaraConfig) (types.Storage, error) {
	var storage types.Storage
	var err error

	enc, err := encryptor.NewEncryptor(conf)
	if err != nil {
		return nil, err
	}

	switch conf.Storage {
	case "vault":
		storage, err = NewVaultService(conf.VaultAddr, conf.VaultToken, conf.DeploymentID)
	case "sql":
		storage, err = NewSqlService(conf.DeploymentID, &conf.DBConfig, enc, conf.LogLevel)
	default:
		return nil, fmt.Errorf("type %s is not supported", conf.Storage)
	}

	if err == nil {
		log.Printf("INF: using %s secret storage", conf.Storage)
	}

	return storage, err
}
