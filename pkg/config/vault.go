package config

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

type VaultService struct {
	data         map[string]interface{}
	appName      string // Used as component Name
	vault        *api.Client
	deploymentID string // Used as vault prefix
}

// NewService instantiate a vault service
func NewService(addr, token, appName string) *VaultService {
	if addr == "" {
		addr = "http://localhost:8200"
	}
	if token == "" {
		panic("VAULT_TOKEN is missing")
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

	s := &VaultService{
		vault:   client,
		appName: appName,
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
	}
	return s
}

func (vs *VaultService) SetDeploymentID(deploymentID string) {
	vs.deploymentID = deploymentID
}

func (vs *VaultService) keyPath(scope string) string {
	return fmt.Sprintf("%s/%s/%s", vs.deploymentID, vs.appName, scope)
}

func (vs *VaultService) transitKeyName() string {
	return fmt.Sprintf("%s_kaigara_%s", vs.deploymentID, vs.appName)
}

func (vs *VaultService) transitKeyExists() (bool, error) {
	secret, err := vs.vault.Logical().Read("transit/keys/" + vs.transitKeyName())
	if err != nil {
		return false, err
	}
	return secret != nil, nil
}

func (vs *VaultService) transitKeyCreate() error {
	_, err := vs.vault.Logical().Write("transit/keys/"+vs.transitKeyName(), map[string]interface{}{
		"force": true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (vs *VaultService) Encrypt(plaintext string) (string, error) {
	secret, err := vs.vault.Logical().Write("transit/encrypt/"+vs.transitKeyName(), map[string]interface{}{
		"plaintext": base64.URLEncoding.EncodeToString([]byte(plaintext)),
	})
	if err != nil {
		return "", err
	}

	ciphertext, ok := secret.Data["ciphertext"]
	if !ok {
		return "", fmt.Errorf("ciphertext not found in vault types.Response")
	}
	return ciphertext.(string), nil
}

// Decrypt the given ciphertext and return the plaintext or an error
func (vs *VaultService) Decrypt(ciphertext string) (string, error) {
	secret, err := vs.vault.Logical().Write("transit/decrypt/"+vs.transitKeyName(), map[string]interface{}{
		"ciphertext": ciphertext,
	})
	if err != nil {
		return "", err
	}

	data, ok := secret.Data["plaintext"]
	if !ok {
		return "", fmt.Errorf("plaintext not found in vault types.Response")
	}

	plaintext, err := base64.URLEncoding.DecodeString(data.(string))
	return string(plaintext), err
}

// LoadSecrets loads existing secrets from vault
func (vs *VaultService) LoadSecrets(scope string) error {
	fmt.Println("Loading secrets...")
	secret, err := vs.vault.Logical().Read(vs.keyPath(scope))
	if err != nil {
		return err
	}
	if secret == nil {
		vs.data = make(map[string]interface{})
	} else {
		vs.data = secret.Data["data"].(map[string]interface{})
	}
	return nil
}

// SetSecret stores all secrets into the memory
func (vs *VaultService) SetSecret(name, value string) error {
	vs.data[name] = value
	return nil
}

// SetSecrets inserts given data into the secret store, overwriting keys if they exist
func (vs *VaultService) SetSecrets(data map[string]interface{}) error {
	for k, v := range data {
		vs.data[k] = v
	}
	return nil
}

// SaveSecrets saves all secrets to a Vault kv secret
func (vs *VaultService) SaveSecrets(scope string) error {
	if vs.deploymentID == "" {
		return fmt.Errorf("Deployment ID is not set, please set deploymentID")
	}

	_, err := vs.vault.Logical().Write(vs.keyPath(scope), map[string]interface{}{
		"data": vs.data,
	})
	if err == nil {
		fmt.Printf("Stored secrets in vault secret: %s\n", vs.keyPath(scope))
	}
	return err
}

// GetSecrets returns all the secrets currently stored in Vault
func (vs *VaultService) GetSecrets() (map[string]interface{}, error) {
	return vs.data, nil
}

// GetSecret returns a secret value by name
func (vs *VaultService) GetSecret(name string) (interface{}, error) {
	return vs.data[name], nil
}
