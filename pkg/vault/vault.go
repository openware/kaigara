package vault

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/iancoleman/strcase"

	"github.com/openware/kaigara/pkg/encryptor/types"
)

// Service contains scoped secret data, Vault client and configuration
type Service struct {
	data         map[string]map[string]interface{}
	metadata     map[string]map[string]interface{}
	vault        *api.Client
	deploymentID string // Used as vault prefix
	encryptor    types.Encryptor
}

// NewService instantiates a Vault service
func NewService(deploymentID string, encryptor types.Encryptor, addr, token string) (*Service, error) {
	if addr == "" {
		addr = "http://localhost:8200"
	}

	if token == "" {
		return nil, fmt.Errorf("KAIGARA_VAULT_TOKEN is missing")
	}

	if deploymentID == "" {
		return nil, fmt.Errorf("KAIGARA_DEPLOYMENT_ID is missing")
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

	s := &Service{
		deploymentID: deploymentID,
		vault:        client,
		encryptor:    encryptor,
	}

	err = s.startRenewToken(token)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (vs *Service) startRenewToken(token string) error {
	secret, err := vs.vault.Auth().Token().Lookup(token)
	if err != nil {
		return err
	}

	var renewable bool
	if v, ok := secret.Data["renewable"]; ok {
		renewable, _ = v.(bool)
	}

	if !renewable {
		return nil
	}

	watcher, err := vs.vault.NewLifetimeWatcher(&api.LifetimeWatcherInput{
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

	go watcher.Start()
	go func() {
		for {
			select {
			case <-watcher.DoneCh():
				return
			case <-watcher.RenewCh():
			}
		}
	}()
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

// LoadSecrets loads existing secrets from vault
func (vs *Service) Read(appName, scope string) error {
	secret, err := vs.vault.Logical().Read(vs.keyPath(appName, scope))
	if err != nil {
		return err
	}

	if vs.data == nil {
		vs.data = make(map[string]map[string]interface{})
	}
	if vs.data[appName] == nil {
		vs.data[appName] = make(map[string]interface{})
	}
	if vs.metadata == nil {
		vs.metadata = make(map[string]map[string]interface{})
	}
	if vs.metadata[appName] == nil {
		vs.metadata[appName] = make(map[string]interface{})
	}

	if secret == nil || secret.Data == nil || secret.Data["data"] == nil {
		vs.data[appName][scope] = make(map[string]interface{})
		vs.metadata[appName][scope] = make(map[string]interface{})
	} else {
		vs.data[appName][scope] = secret.Data["data"].(map[string]interface{})
		rawMetadata := secret.Data["metadata"]
		if rawMetadata == nil {
			return fmt.Errorf("metadata not found, make sure you have enabled KV v2 enabled: vault secrets enable -version=2 -path=secret kv")
		}
		vs.metadata[appName][scope] = rawMetadata.(map[string]interface{})
	}

	return nil
}

// SetEntry stores all secrets into the memory
func (vs *Service) SetEntry(appName, scope, name string, value interface{}) error {
	if scope == "secret" {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid value for %s, must be a string: %v", name, value)
		}

		encrypted, err := vs.encryptor.Encrypt(str, vs.transitKeyName(appName))
		if err != nil {
			return err
		}

		vs.data[appName][scope].(map[string]interface{})[name] = encrypted
	} else {
		vs.data[appName][scope].(map[string]interface{})[name] = value
	}

	return nil
}

// SetEntries inserts given data into the secret store, overwriting keys if they exist
func (vs *Service) SetEntries(appName, scope string, data map[string]interface{}) error {
	for k, v := range data {
		err := vs.SetEntry(appName, scope, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Write saves all secrets to a Vault kv secret
func (vs *Service) Write(appName, scope string) error {
	if vs.deploymentID == "" {
		return fmt.Errorf("Deployment ID is not set, please set deploymentID")
	}

	metadata, err := vs.vault.Logical().Write(vs.keyPath(appName, scope), map[string]interface{}{
		"data": vs.data[appName][scope],
	})
	if err == nil {
		vs.metadata[appName][scope] = metadata.Data
	}
	return err
}

// GetEntries returns all the secrets currently stored in Vault
func (vs *Service) GetEntries(appName, scope string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for k := range vs.data[appName][scope].(map[string]interface{}) {
		val, err := vs.GetEntry(appName, scope, k)
		if err != nil {
			return nil, err
		}

		res[k] = val
	}
	return res, nil
}

// GetEntry returns a secret value by name
func (vs *Service) GetEntry(appName, scope, name string) (interface{}, error) {
	// Since secret scope only supports strings, return a decrypted string
	if scope == "secret" {
		scopeSecrets, ok := vs.data[appName][scope].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("scope '%s' is not a map", scope)
		}

		rawValue, ok := scopeSecrets[name]
		if !ok {
			return nil, nil
		}

		str, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for %s, must be a string: %v", name, rawValue)
		}

		decrypted, err := vs.encryptor.Decrypt(str, vs.transitKeyName(appName))
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return vs.data[appName][scope].(map[string]interface{})[name], nil
}

// ListEntries returns a slice containing all secret keys of a scope
func (vs *Service) ListEntries(appName, scope string) ([]string, error) {
	secrets := vs.data[appName][scope].(map[string]interface{})
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

func (vs *Service) PushPolicies(policies map[string]string) error {
	err := vs.Read("tokens", "secret")
	if err != nil {
		return err
	}

	for component, rules := range policies {
		name := fmt.Sprintf("%s_%s", vs.deploymentID, component)
		fmt.Println("Loading policy", name)
		err := vs.vault.Sys().PutPolicy(name, rules)
		if err != nil {
			return err
		}

		fmt.Println("Creating token", name)
		t := true
		token, err := vs.vault.Auth().Token().Create(&api.TokenCreateRequest{
			Policies:  []string{name},
			Renewable: &t,
			TTL:       "240h",
			Period:    "240h",
		})
		if err != nil {
			return err
		}

		keyName := strcase.ToLowerCamel(component + "_vault_token")
		if err := vs.SetEntry("tokens", keyName, token.Auth.ClientToken, "secret"); err != nil {
			return err
		}
	}
	return nil
}

// GetCurrentVersion returns current data version in cache
func (vs *Service) GetCurrentVersion(appName, scope string) (int64, error) {
	var versionNumber int64 = -1
	v := vs.metadata[appName][scope].(map[string]interface{})["version"]
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
	if err != nil || metadata == nil {
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

// Delete key from Data, Metadata and Vault
func (vs *Service) DeleteEntry(appName, scope, name string) error {
	metadata, err := vs.vault.Logical().Delete(vs.keyPath(appName, scope))
	if err != nil {
		return err
	}
	if metadata != nil {
		vs.metadata[appName][scope] = metadata.Data
	}
	delete(vs.data[appName][scope].(map[string]interface{}), name)
	err = vs.Write(appName, scope)
	return err
}
