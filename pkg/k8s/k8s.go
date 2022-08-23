package k8s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openware/kaigara/pkg/encryptor/types"
	"github.com/openware/pkg/kube"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Service contains a K8s client
type Service struct {
	client       *kube.K8sClient
	deploymentID string
	ds           map[string]map[string]map[string]interface{}
	encryptor    types.Encryptor
}

func NewService(deploymentID string, client *kube.K8sClient, encryptor types.Encryptor) (*Service, error) {
	return &Service{
		client:       client,
		deploymentID: deploymentID,
		encryptor:    encryptor,
	}, nil
}

func secretName(appName, scope string) string {
	return fmt.Sprintf("kaigara-%s-%s", appName, scope)
}

func (ss *Service) namespace() string {
	return strings.ReplaceAll(ss.deploymentID, "_", "-")
}

func (ss *Service) Read(appName, scope string) error {
	secretName := secretName(appName, scope)
	val := make(map[string]interface{})
	val["version"] = int64(0)

	secrets, err := ss.client.ReadSecret(secretName, ss.namespace())
	if err == nil {
		for name, secret := range secrets {
			data := string(secret)
			if name == "version" {
				if ver, err := strconv.ParseInt(data, 10, 64); err == nil {
					val[name] = ver
				}
			} else {
				val[name] = data
			}
		}
	} else {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	if ss.ds == nil {
		ss.ds = make(map[string]map[string]map[string]interface{})
	}
	if ss.ds[appName] == nil {
		ss.ds[appName] = make(map[string]map[string]interface{})
	}
	ss.ds[appName][scope] = val

	return nil
}

func (ss *Service) Write(appName, scope string) error {
	// verify data stored in secret store
	val, ok := ss.ds[appName][scope]
	if !ok {
		return fmt.Errorf("scope '%s' in '%s' app is: %v", scope, appName, val)
	}

	secretName := secretName(appName, scope)
	namespace := ss.namespace()

	secrets, err := ss.client.ReadSecret(secretName, namespace)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	// update version from exiting secret, create version = 0 for a new secret
	newVersion := int64(0)
	if version, ok := secrets["version"]; ok {
		ver, err := strconv.ParseInt(string(version), 0, 16)
		if err != nil {
			return err
		}
		newVersion = ver + 1
	}
	val["version"] = newVersion

	// delete and re-create a K8s secret with a secretName from a val
	if err := ss.client.DeleteSecret(secretName, namespace); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	if err := ss.client.UpdateSecret(secretName, namespace, val); err != nil {
		return err
	}

	return nil
}

func (ss *Service) SetEntry(appName, scope, name string, value interface{}) error {
	if scope == "secret" && name != "version" {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid value for %s, must be a string: %v", name, value)
		}
		encrypted, err := ss.encryptor.Encrypt(str, appName)
		if err != nil {
			return err
		}

		ss.ds[appName][scope][name] = encrypted
	} else {
		ss.ds[appName][scope][name] = value
	}

	return nil
}

func (ss *Service) SetEntries(appName, scope string, data map[string]interface{}) error {
	for k, v := range data {
		err := ss.SetEntry(appName, scope, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ss *Service) GetEntry(appName, scope, name string) (interface{}, error) {
	// Since secret scope only supports strings, return a decrypted string
	scopeSecrets, ok := ss.ds[appName][scope]
	if !ok {
		return nil, fmt.Errorf("scope '%s' is not loaded", scope)
	}

	if scope == "secret" && name != "version" {
		rawValue, ok := scopeSecrets[name]
		if !ok {
			return nil, nil
		}

		str, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for %s, must be a string: %v", name, rawValue)
		}

		decrypted, err := ss.encryptor.Decrypt(str, appName)
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return ss.ds[appName][scope][name], nil
}

func (ss *Service) GetEntries(appName, scope string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for k := range ss.ds[appName][scope] {
		val, err := ss.GetEntry(appName, scope, k)
		if err != nil {
			return nil, err
		}

		res[k] = val
	}
	return res, nil
}

func (ss *Service) ListEntries(appName, scope string) ([]string, error) {
	val, ok := ss.ds[appName][scope]
	if !ok {
		return []string{}, nil
	}

	res := make([]string, len(val))
	i := 0
	for k := range val {
		res[i] = k
		i++
	}

	return res, nil
}

func (ss *Service) DeleteEntry(appName, scope, name string) error {
	delete(ss.ds[appName][scope], name)

	return nil
}

func (ss *Service) ListAppNames() ([]string, error) {
	secrets, err := ss.client.GetSecrets(ss.namespace())
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, secret := range secrets {
		name := secret.Name
		// secret name format is kaigara-${app_name}-${scope}
		if strings.HasPrefix(name, "kaigara") {
			apps := strings.Split(name, "-")
			appName := apps[1]

			found := false
			for _, app := range names {
				if app == appName {
					found = true
				}
			}
			if !found {
				names = append(names, appName)
			}
		}
	}

	return names, nil
}

func (ss *Service) GetCurrentVersion(appName, scope string) (int64, error) {
	if ss.ds[appName][scope] == nil {
		return 0, fmt.Errorf("failed to get %s.%s.version: scope is not loaded", appName, scope)
	}

	res, ok := ss.ds[appName][scope]["version"].(int64)
	if !ok {
		return 0, fmt.Errorf("failed to get %s.%s.version: type assertion to int64 failed, actual value: %v", appName, scope, res)
	}

	return res, nil
}

func (ss *Service) GetLatestVersion(appName, scope string) (int64, error) {
	secretName := secretName(appName, scope)

	secrets, err := ss.client.ReadSecret(secretName, ss.namespace())
	if err != nil {
		if errors.IsNotFound(err) {
			ver, err := ss.GetCurrentVersion(appName, scope)
			if err != nil {
				return 0, err
			}

			return ver, nil
		}

		return 0, fmt.Errorf("failed to check for an existing value in the kubernetes: %s", err)
	}

	data := string(secrets["version"])
	ver, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return ver, nil
}
