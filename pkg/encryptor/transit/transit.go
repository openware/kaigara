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
	vault *api.Client
}

// NewVaultEncryptor instantiate a vault encryption service
func NewVaultEncryptor(addr, token string) (*VaultEncryptor, error) {
	if addr == "" {
		addr = "http://localhost:8200"
	}

	if token == "" {
		return nil, fmt.Errorf("vault token is empty")
	}

	config := &api.Config{
		Address: addr,
		Timeout: time.Second * 2,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	s := &VaultEncryptor{
		vault: client,
	}

	if err := s.startRenewToken(token); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *VaultEncryptor) transitKeyExists(appName string) (bool, error) {
	secret, err := s.vault.Logical().Read("transit/keys/" + appName)
	if err != nil {
		return false, err
	}

	return secret != nil, nil
}

func (s *VaultEncryptor) transitKeyCreate(appName string) error {
	_, err := s.vault.Logical().Write("transit/keys/"+appName, map[string]interface{}{
		"force": true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *VaultEncryptor) createTransitKeyIfNotExist(appName string) error {
	ok, err := s.transitKeyExists(appName)
	if err != nil {
		return err
	}

	if !ok {
		err = s.transitKeyCreate(appName)
		if err != nil {
			return err
		}
		log.Println("INFO: Transit key created")
	}

	return nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (s *VaultEncryptor) Encrypt(plaintext, appName string) (string, error) {
	err := s.createTransitKeyIfNotExist(appName)
	if err != nil {
		return "", err
	}

	secret, err := s.vault.Logical().Write("transit/encrypt/"+appName, map[string]interface{}{
		"plaintext": base64.URLEncoding.EncodeToString([]byte(plaintext)),
	})
	if err != nil {
		return "", err
	}

	ciphertext, ok := secret.Data["ciphertext"]
	if !ok {
		return "", fmt.Errorf("ciphertext not found in Vault response")
	}
	return ciphertext.(string), nil
}

// Decrypt the given ciphertext and return the plaintext or an error
func (s *VaultEncryptor) Decrypt(ciphertext, appName string) (string, error) {
	if err := s.createTransitKeyIfNotExist(appName); err != nil {
		return "", err
	}

	secret, err := s.vault.Logical().Write("transit/decrypt/"+appName, map[string]interface{}{
		"ciphertext": ciphertext,
	})
	if err != nil {
		return "", err
	}

	data, ok := secret.Data["plaintext"]
	if !ok {
		return "", fmt.Errorf("plaintext not found in Vault response")
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
		log.Println("WRN: token is not renewable")
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

	log.Println("INF: launching Vault token renewal")
	go watcher.Start()
	go func() {
		for {
			select {
			case err := <-watcher.DoneCh():
				log.Printf("ERR: token renew failed: %s\n", err.Error())
				return
			case <-watcher.RenewCh():
				log.Println("INF: successfully renewed token")
			}
		}
	}()
	return nil
}
