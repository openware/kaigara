package config

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	client *api.Client
	path   string
}

func NewVaultConfig(addr, token, path string) *VaultConfig {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	client, err := api.NewClient(&api.Config{Address: addr, HttpClient: httpClient})
	if err != nil {
		panic(err)
	}
	client.SetToken(token)
	return &VaultConfig{
		client: client,
		path:   path,
	}
}

func (vc *VaultConfig) ListEntries() map[string]interface{} {
	secret, err := vc.client.Logical().Read(path.Join("secret/data", vc.path))
	if err != nil {
		panic(err)
	}

	if secret == nil || secret.Data == nil {
		return map[string]interface{}{}
	}

	data, ok := secret.Data["data"]
	if !ok || data == nil {
		return map[string]interface{}{}
	}
	return data.(map[string]interface{})
}

// Service implements the EncryptionService interface using vault transit key
type Service struct {
	data         map[string]interface{}
	appName      string
	vault        *api.Client
	deploymentID string // Used as vault prefix
}

// NewService instantiate a vault encryption service
func NewService(addr, token, appName string) *Service {
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

	s := &Service{
		vault:   client,
		appName: appName,
	}
	return s
}

// SetDeploymentID set the vault secret kv engine mount point
func (vs *Service) SetDeploymentID(mount string) {
	vs.deploymentID = mount
}

func (vs *Service) keyPath(scope string) string {
	return fmt.Sprintf("%s/%s/%s", vs.deploymentID, vs.appName, scope)
}

// LoadSecrets loads existing secrets from vault
func (vs *Service) LoadSecrets(scope string) error {
	fmt.Println("Loading secrets...")
	secret, err := vs.vault.Logical().Read(vs.keyPath(scope))
	if err != nil {
		return err
	}
	if secret == nil {
		vs.data[scope] = make(map[string]interface{})
	} else {
		vs.data[scope] = secret.Data["data"].(map[string]interface{})
	}
	return nil
}

// SetSecret stores all secrets into the memory
func (vs *Service) SetSecret(name, value, scope string) error {
	vs.data[scope][name] = value
	return nil
}

// SetSecrets inserts given data into the secret store, overwriting keys if they exist
func (vs *Service) SetSecrets(data map[string]interface{}) error {
	for k, v := range data {
		vs.data[k] = v
	}
	return nil
}

// SaveSecrets saves all secrets to a Vault kv secret
func (vs *Service) SaveSecrets() error {
	if vs.secretMount == "" {
		return fmt.Errorf("Secret mount is empty")
	}
	if vs.deploymentID == "" {
		return fmt.Errorf("Deployment ID is not set, please set deploymentID")
	}

	_, err := vs.vault.Logical().Write(vs.keyPath("secrets"), map[string]interface{}{
		"data": vs.data,
	})
	if err == nil {
		fmt.Printf("Stored secrets in vault secret: %s\n", vs.keyPath("secrets"))
	}
	return err
}

// GetSecrets returns all the secrets currently stored in Vault
func (vs *Service) GetSecrets() (map[string]interface{}, error) {
	return vs.data, nil
}

// GetSecret returns a secret value by name
func (vs *Service) GetSecret(name string) (interface{}, error) {
	return vs.data[name], nil
}
