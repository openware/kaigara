package vault

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

type Service struct {
	data         map[string]interface{}
	appName      string // Used as component Name
	vault        *api.Client
	deploymentID string // Used as vault prefix
}

// NewService instantiate a vault service
func NewService(addr, token, appName, deploymentID string) *Service {
	if addr == "" {
		addr = "http://localhost:8200"
	}
	if token == "" {
		panic("KAIGARA_VAULT_TOKEN is missing")
	}

	if appName == "" {
		panic("KAIGARA_APP_NAME is missing")
	}

	if deploymentID == "" {
		panic("KAIGARA_DEPLOYMENT_ID is missing")
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

	s := &Service{
		vault:        client,
		appName:      appName,
		deploymentID: deploymentID,
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

func (vs *Service) keyPath(scope string) string {
	return fmt.Sprintf("kv/%s/%s/%s", vs.deploymentID, vs.appName, scope)
}

func (vs *Service) transitKeyName() string {
	return fmt.Sprintf("%s_kaigara_%s", vs.deploymentID, vs.appName)
}

func (vs *Service) transitKeyExists() (bool, error) {
	secret, err := vs.vault.Logical().Read("transit/keys/" + vs.transitKeyName())
	if err != nil {
		return false, err
	}
	return secret != nil, nil
}

func (vs *Service) transitKeyCreate() error {
	_, err := vs.vault.Logical().Write("transit/keys/"+vs.transitKeyName(), map[string]interface{}{
		"force": true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (vs *Service) Encrypt(plaintext string) (string, error) {
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
func (vs *Service) Decrypt(ciphertext string) (string, error) {
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
func (vs *Service) LoadSecrets(scope string) error {
	fmt.Println("Loading secrets...")
	secret, err := vs.vault.Logical().Read(vs.keyPath(scope))
	if err != nil {
		return err
	}

	// fmt.Println("Secret:", secret)

	if vs.data == nil {
		vs.data = make(map[string]interface{})
	}

	if secret == nil || secret.Data == nil {
		// fmt.Println("Secret or data is nil, overwriting:", vs.data[scope])
		vs.data[scope] = make(map[string]interface{})
	} else {
		// fmt.Println("Secret is not nil, writing", secret.Data["data"])
		vs.data[scope] = secret.Data["data"].(map[string]interface{})
	}

	return nil
}

// SetSecret stores all secrets into the memory
func (vs *Service) SetSecret(name string, value interface{}, scope string) error {
	// Secret scope only supports strings
	if scope == "secret" {
		encrypted, err := vs.Encrypt(fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}

		vs.data[scope].(map[string]interface{})[name] = encrypted
	} else {
		vs.data[scope].(map[string]interface{})[name] = value
	}

	return nil
}

// SetSecrets inserts given data into the secret store, overwriting keys if they exist
func (vs *Service) SetSecrets(data map[string]interface{}, scope string) error {
	for k, v := range data {
		vs.SetSecret(k, v, scope)
	}
	return nil
}

// SaveSecrets saves all secrets to a Vault kv secret
// TODO should save data from all scopes
func (vs *Service) SaveSecrets(scope string) error {
	if vs.deploymentID == "" {
		return fmt.Errorf("Deployment ID is not set, please set deploymentID")
	}

	_, err := vs.vault.Logical().Write(vs.keyPath(scope), map[string]interface{}{
		"data": vs.data[scope],
	})
	if err == nil {
		fmt.Printf("Stored secrets in vault secret: %s\n", vs.keyPath(scope))
	}
	return err
}

// GetSecrets returns all the secrets currently stored in Vault
func (vs *Service) GetSecrets(scope string) (map[string]interface{}, error) {
	return vs.data[scope].(map[string]interface{}), nil
}

// GetSecret returns a secret value by name
func (vs *Service) GetSecret(name, scope string) (interface{}, error) {
	// Since secret scope only supports strings, return a decrypted string
	if scope == "secret" {
		decrypted, err := vs.Decrypt(fmt.Sprintf("%v", vs.data[scope].(map[string]interface{})[name]))
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return vs.data[scope].(map[string]interface{})[name], nil
}
