package storage

import (
	"fmt"
	"log"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/encryptor/aes"
	"github.com/openware/kaigara/pkg/encryptor/plaintext"
	"github.com/openware/kaigara/pkg/encryptor/transit"
	enc "github.com/openware/kaigara/pkg/encryptor/types"
	"github.com/openware/kaigara/pkg/sql"
	"github.com/openware/kaigara/pkg/vault"
	"github.com/openware/kaigara/types"
)

func GetStorageService(conf *config.KaigaraConfig) (types.Storage, error) {
	var storage types.Storage
	var err error

	enc, err := NewEncryptor(conf)
	if err != nil {
		return nil, err
	}

	switch conf.Storage {
	case "vault":
		storage, err = vault.NewService(conf.VaultAddr, conf.VaultToken, conf.DeploymentID)
	case "sql":
		storage, err = sql.NewService(conf.DeploymentID, &conf.DBConfig, enc, conf.LogLevel)
	default:
		return nil, fmt.Errorf("type %s is not supported", conf.Storage)
	}

	if err == nil {
		log.Printf("INF: using %s secret storage", conf.Storage)
	}

	return storage, err
}

func NewEncryptor(conf *config.KaigaraConfig) (enc.Encryptor, error) {
	var enc enc.Encryptor
	var err error

	switch conf.EncryptMethod {
	case "transit":
		enc, err = transit.NewVaultEncryptor(conf.VaultAddr, conf.VaultToken)
		if err == nil {
			log.Println("INF: starting vault transit secret engine encryption!")
		}

	case "aes":
		enc, err = aes.NewAESEncryptor([]byte(conf.AesKey))
		if err == nil {
			log.Println("INF: starting in-memory encryption!")
		}

	case "plaintext":
		enc = plaintext.NewPlaintextEncryptor()
		log.Println("INF: starting plaintext encryption (default)")

	default:
		return nil, fmt.Errorf("type '%s' is not supported", conf.EncryptMethod)
	}

	return enc, err
}
