package storage

import (
	"fmt"
	"log"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/pkg/encryptor/aes"
	"github.com/openware/pkg/encryptor/plaintext"
	"github.com/openware/pkg/encryptor/transit"
	enc "github.com/openware/pkg/encryptor/types"
	"github.com/openware/kaigara/pkg/k8s"
	"github.com/openware/kaigara/pkg/sql"
	"github.com/openware/pkg/vault"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/kube"
	"k8s.io/client-go/tools/clientcmd"
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
		storage, err = vault.NewService(conf.DeploymentID, enc, conf.VaultAddr, conf.VaultToken)
	case "sql":
		storage, err = sql.NewService(conf.DeploymentID, &conf.DBConfig, enc, conf.LogLevel)
	case "k8s":
		// create a new client from kubeconfig
		config, cfgErr := clientcmd.BuildConfigFromFlags("", conf.KubeConfig)
		if cfgErr != nil {
			return nil, cfgErr
		}

		client, cfgErr := kube.NewClient(config)
		if cfgErr != nil {
			return nil, cfgErr
		}

		storage, err = k8s.NewService(conf.DeploymentID, client, enc)
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

func CleanAll(ss types.Storage, appNames []string, scopes []string) error {
	if !strings.Contains(strings.Join(appNames, ","), "global") {
		appNames = append(appNames, "global")
	}

	for _, appName := range appNames {
		for _, scope := range scopes {
			if err := ss.Read(appName, scope); err != nil {
				return err
			}

			entries, err := ss.ListEntries(appName, scope)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if err := ss.DeleteEntry(appName, scope, entry); err != nil {
					return err
				}
			}

			if err := ss.Write(appName, scope); err != nil {
				return err
			}
		}
	}

	return nil
}
