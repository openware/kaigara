package config

import (
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
