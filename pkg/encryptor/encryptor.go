package encryptor

import (
	"log"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
)

func NewEncryptor(conf *config.KaigaraConfig) (types.Encryptor, error) {
	var encryptor types.Encryptor
	var err error

	switch conf.EncryptMethod {
	case "transit":
		encryptor, err = NewVaultEncryptor(conf.VaultAddr, conf.VaultToken)
		if err == nil {
			log.Println("INF: starting vault transit secret engine encryption!")
		}

	case "aes":
		encryptor, err = NewAESEncryptor([]byte(conf.AesKey))
		if err == nil {
			log.Println("INF: starting in-memory encryption!")
		}

	default:
		encryptor = NewPlaintextEncryptor()
		log.Println("INF: starting plaintext encryption (default)")
	}

	return encryptor, err
}
