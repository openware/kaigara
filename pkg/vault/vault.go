package vault

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

// Service contains scoped secret data, Vault client and configuration
type Service struct {
	data         map[string]interface{}
	metadata     map[string]interface{}
	vault        *api.Client
	deploymentID string // Used as vault prefix
}

// NewService instantiates a Vault service
func NewService(addr, token, deploymentID string) *Service {
	if addr == "" {
		addr = "http://localhost:8200"
	}
	if token == "" {
		panic("KAIGARA_VAULT_TOKEN is missing")
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
		deploymentID: deploymentID,
	}

	return s
}

func (vs *Service) initTransitKey(appName string) error {
	ok, err := vs.transitKeyExists(appName)
	if err != nil {
		return err
	}

	if !ok {
		err = vs.transitKeyCreate(appName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vs *Service) secretPath(appName, directory, scope string) string {
	return fmt.Sprintf("secret/%s/%s/%s/%s", directory, vs.deploymentID, appName, scope)
}

func (vs *Service) keyPath(appName, scope string) string {
	return vs.secretPath(appName, "data", scope)
}

func (vs *Service) metadataPath(appName, scope string) string {
	return vs.secretPath(appName, "metadata", scope)
}

func (vs *Service) transitKeyName(appName string) string {
	return fmt.Sprintf("%s_kaigara_%s", vs.deploymentID, appName)
}

func (vs *Service) transitKeyExists(appName string) (bool, error) {
	secret, err := vs.vault.Logical().Read("transit/keys/" + vs.transitKeyName(appName))
	if err != nil {
		return false, err
	}
	return secret != nil, nil
}

func (vs *Service) transitKeyCreate(appName string) error {
	_, err := vs.vault.Logical().Write("transit/keys/"+vs.transitKeyName(appName), map[string]interface{}{
		"force": true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Encrypt the plaintext argument and return a ciphertext string or an error
func (vs *Service) Encrypt(appName, plaintext string) (string, error) {
	secret, err := vs.vault.Logical().Write("transit/encrypt/"+vs.transitKeyName(appName), map[string]interface{}{
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
func (vs *Service) Decrypt(appName, ciphertext string) (string, error) {
	secret, err := vs.vault.Logical().Write("transit/decrypt/"+vs.transitKeyName(appName), map[string]interface{}{
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
func (vs *Service) LoadSecrets(appName, scope string) error {
	fmt.Println("Loading secrets...")
	secret, err := vs.vault.Logical().Read(vs.keyPath(appName, scope))
	if err != nil {
		return err
	}

	if vs.data == nil {
		vs.data = make(map[string]interface{})
	}

	if vs.metadata == nil {
		vs.metadata = make(map[string]interface{})
	}

	if secret == nil || secret.Data == nil || secret.Data["data"] == nil {
		vs.data[scope] = make(map[string]interface{})
		vs.metadata[scope] = make(map[string]interface{})
	} else {
		vs.data[scope] = secret.Data["data"].(map[string]interface{})
		vs.metadata[scope] = secret.Data["metadata"].(map[string]interface{})
	}

	return nil
}

// SetSecret stores all secrets into the memory
func (vs *Service) SetSecret(appName, name string, value interface{}, scope string) error {
	if scope == "secret" {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("secretStore.SetSecret: %s is not a string", name)
		}

		encrypted, err := vs.Encrypt(appName, str)
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
func (vs *Service) SetSecrets(appName string, data map[string]interface{}, scope string) error {
	for k, v := range data {
		vs.SetSecret(appName, k, v, scope)
	}
	return nil
}

// SaveSecrets saves all secrets to a Vault kv secret
func (vs *Service) SaveSecrets(appName, scope string) error {
	if vs.deploymentID == "" {
		return fmt.Errorf("Deployment ID is not set, please set deploymentID")
	}

	metadata, err := vs.vault.Logical().Write(vs.keyPath(appName, scope), map[string]interface{}{
		"data": vs.data[scope],
	})
	if err == nil {
		fmt.Printf("Stored secrets in vault secret: %s\n", vs.keyPath(appName, scope))
		vs.metadata[scope] = metadata.Data
	}
	return err
}

// GetSecrets returns all the secrets currently stored in Vault
func (vs *Service) GetSecrets(appName, scope string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for k := range vs.data[scope].(map[string]interface{}) {
		val, err := vs.GetSecret(appName, k, scope)
		if err != nil {
			return nil, err
		}

		res[k] = val
	}
	return res, nil
}

// GetSecret returns a secret value by name
func (vs *Service) GetSecret(appName, name, scope string) (interface{}, error) {
	// Since secret scope only supports strings, return a decrypted string
	if scope == "secret" {
		str, ok := vs.data[scope].(map[string]interface{})[name].(string)
		if !ok {
			return nil, fmt.Errorf("secretStore.GetSecret: %s is not a string", name)
		}

		decrypted, err := vs.Decrypt(appName, str)
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return vs.data[scope].(map[string]interface{})[name], nil
}

// ListSecrets returns a slice containing all secret keys of a scope
func (vs *Service) ListSecrets(scope string) ([]string, error) {
	secrets := vs.data[scope].(map[string]interface{})
	keys := make([]string, len(secrets))

	i := 0
	for k := range secrets {
		keys[i] = k
		i++
	}

	return keys, nil
}

// ListAppNames returns a slice containing all app names inside the deploymentID namespace
func (vs *Service) ListAppNames() ([]string, error) {
	secret, err := vs.vault.Logical().List(fmt.Sprintf("secret/metadata/%s", vs.deploymentID))
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	var res []string
	secretKeys := secret.Data["keys"].([]interface{})

	for _, val := range secretKeys {
		res = append(res, strings.ReplaceAll(val.(string), "/", ""))
	}

	return res, nil
}

// GetCurrentVersion returns current data version in cache
func (vs *Service) GetCurrentVersion(scope string) (int64, error) {
	var versionNumber int64 = -1
	v := vs.metadata[scope].(map[string]interface{})["version"]
	if v != nil {
		version, err := v.(json.Number).Int64()
		if err != nil {
			return versionNumber, err
		}
		versionNumber = version
	}
	return versionNumber, nil
}

// GetLatestVersion returns latest data version from vault
func (vs *Service) GetLatestVersion(appName, scope string) (int64, error) {
	var versionNumber int64 = -1
	metadata, err := vs.vault.Logical().Read(vs.metadataPath(appName, scope))
	if err != nil {
		return versionNumber, err
	}

	v := metadata.Data["current_version"]
	if v != nil {
		version, err := v.(json.Number).Int64()
		if err != nil {
			return versionNumber, err
		}
		versionNumber = version
	}
	return versionNumber, nil
}
