package transit

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultEncryptor implements Encryptor interface by using Vault transit
type VaultEncryptor struct {
	vault   *api.Client
	appName string
}

// NewVaultEncryptor instantiate a vault encryption service
func NewVaultEncryptor(addr, token, appName string) *VaultEncryptor {
	if addr == "" {
		addr = "http://localhost:8200"
	}
	if token == "" {
		panic("VAULT_TOKEN is missing")
	}
	if appName == "" {
		panic("VAULT_APP_NAME is missing")
	}

	config := &api.Config{
		Address: addr,
		Timeout: time.Second * 2,
	}

	client, err := api.NewClient(config)
	if err != nil {
		panic(err)
	}
	client.SetToken(token)

	s := &VaultEncryptor{
		appName: appName,
		vault:   client,
	}

	ok, err := s.transitKeyExists()
	if err != nil {
		panic(err)
	}
	if !ok {
		err = s.transitKeyCreate()
		if err != nil {
			panic(err)
		}
		log.Println("INFO: Transit key created")
	}

	err = s.startRenewToken(token)
	if err != nil {
		panic(err)
	}

	return s
}

func (s *VaultEncryptor) transitKeyExists() (bool, error) {
	secret, err := s.vault.Logical().Read("transit/keys/" + s.appName)
	if err != nil {
		return false, err
	}
	return secret != nil, nil
}

func (s *VaultEncryptor) transitKeyCreate() error {
	_, err := s.vault.Logical().Write("transit/keys/"+s.appName, map[string]interface{}{
		"force": true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (s *VaultEncryptor) Encrypt(plaintext string) (string, error) {
	secret, err := s.vault.Logical().Write("transit/encrypt/"+s.appName, map[string]interface{}{
		"plaintext": base64.URLEncoding.EncodeToString([]byte(plaintext)),
	})
	if err != nil {
		return "", err
	}

	ciphertext, ok := secret.Data["ciphertext"]
	if !ok {
		return "", fmt.Errorf("ciphertext not found in vault response")
	}
	return ciphertext.(string), nil
}

// Decrypt the given ciphertext and return the plaintext or an error
func (s *VaultEncryptor) Decrypt(ciphertext string) (string, error) {
	secret, err := s.vault.Logical().Write("transit/decrypt/"+s.appName, map[string]interface{}{
		"ciphertext": ciphertext,
	})
	if err != nil {
		return "", err
	}

	data, ok := secret.Data["plaintext"]
	if !ok {
		return "", fmt.Errorf("plaintext not found in vault response")
	}

	plaintext, err := base64.URLEncoding.DecodeString(data.(string))
	return string(plaintext), err
}

func (s *VaultEncryptor) startRenewToken(token string) error {
	secret, err := s.vault.Auth().Token().Lookup(token)
	if err != nil {
		return err
	}

	var renewable bool
	if v, ok := secret.Data["renewable"]; ok {
		renewable, _ = v.(bool)
	}

	if !renewable {
		log.Println("WARN: token is not renewable")
		return nil
	}

	watcher, err := s.vault.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret: &api.Secret{
			Auth: &api.SecretAuth{
				Renewable:   renewable,
				ClientToken: token,
			},
		},
	})

	if err != nil {
		return err
	}

	log.Println("INFO: launching token renewal")
	go watcher.Start()
	go func() {
		for {
			select {
			case err := <-watcher.DoneCh():
				log.Printf("ERROR: Token renew failed: %s\n", err.Error())
				return
			case <-watcher.RenewCh():
				log.Println("INFO: Successfully renewed token")
			}
		}
	}()
	return nil
}
