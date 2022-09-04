package k8s

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/pkg/encryptor/aes"
	"github.com/openware/pkg/encryptor/plaintext"
	"github.com/openware/pkg/encryptor/types"
	"github.com/openware/pkg/kube"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var deploymentID = "odax"
var appNames = []string{"finex", "storage"}
var scopes = []string{"public", "private", "secret"}
var encryptors map[string]types.Encryptor
var data map[string][]byte
var secrets map[string]*v1.Secret
var client = &kube.K8sClient{
	Client: fake.NewSimpleClientset(),
}

func TestMain(m *testing.M) {
	aesEncrypt, err := aes.NewAESEncryptor([]byte("1234567890123456"))
	if err != nil {
		panic(err)
	}
	plainEncrypt := plaintext.NewPlaintextEncryptor()

	encryptors = map[string]types.Encryptor{
		"aes":       aesEncrypt,
		"plaintext": plainEncrypt,
	}

	data = map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	secrets = map[string]*v1.Secret{
		"finex-public":    MockSecret("kaigara-finex-public", mockNamespace, data),
		"finex-private":   MockSecret("kaigara-finex-private", mockNamespace, data),
		"finex-secret":    MockSecret("kaigara-finex-secret", mockNamespace, data),
		"storage-public":  MockSecret("kaigara-storage-public", mockNamespace, data),
		"storage-private": MockSecret("kaigara-storage-private", mockNamespace, data),
		"storage-secret":  MockSecret("kaigara-storage-secret", mockNamespace, data),
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestRead(t *testing.T) {
	for _, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		ss.client = NewMockClient(
			secrets["finex-public"],
			secrets["finex-private"],
			secrets["finex-secret"],
			secrets["storage-public"],
			secrets["storage-private"],
			secrets["storage-secret"],
		)

		for _, appName := range appNames {
			for _, scope := range scopes {
				err := ss.Read(appName, scope)
				assert.NoError(t, err)

				if _, ok := ss.ds[appName]; !ok {
					assert.Fail(t, fmt.Sprintf("fail to read app %s", appName))
				}
			}
		}

		assert.NoError(t, err)
	}
}

func TestWrite(t *testing.T) {
	for _, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		ss.client = NewMockClient()
		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			for key, val := range data {
				ss.ds[appName][key] = val
			}

			err := ss.Write(appName, "secret")
			assert.NoError(t, err)

			// check version should be assigned
			res := ss.ds[appName]["version"].(int64)
			assert.Equal(t, int64(0), res)
		}
	}
}

func TestWrite_UpdateVersion(t *testing.T) {
	for _, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		ss.client = NewMockClient()
		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			for key, val := range data {
				ss.ds[appName][key] = val
			}

			for i := 0; i < 3; i++ {
				err := ss.Write(appName, "secret")
				assert.NoError(t, err)
			}

			// check version should be increased
			res := ss.ds[appName]["version"].(int64)
			assert.Equal(t, int64(2), res)
		}
	}
}

func TestSetEntry(t *testing.T) {
	for encrypt, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		ss.client = NewMockClient()
		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			for _, scope := range scopes {
				for key, val := range data {
					err := ss.SetEntry(appName, scope, key, string(val))
					assert.NoError(t, err)
				}
			}
		}

		for _, appName := range appNames {
			isEncoded := false
			if encrypt != "plaintext" {
				isEncoded = true
			}

			for key, val := range ss.ds[appName] {
				if isEncoded {
					assert.NotEqual(t, string(data[key]), val)
				} else {
					assert.Equal(t, string(data[key]), val)
				}
			}
		}
	}
}

func TestGetEntry(t *testing.T) {
	for encrypt, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			isEncoded := false
			if encrypt != "plaintext" {
				isEncoded = true
			}

			for key, val := range data {
				encoded := string(val)
				if isEncoded {
					encoded, err = encryptor.Encrypt(string(val), appName)
					assert.NoError(t, err)
				}

				ss.ds[appName][key] = encoded
			}
		}

		for _, appName := range appNames {
			for _, scope := range scopes {
				for key, val := range data {
					entry, err := ss.GetEntry(appName, scope, key)
					assert.NoError(t, err)

					assert.Equal(t, string(val), entry.(string))
				}
			}
		}
	}
}

func TestListEntries(t *testing.T) {
	for encrypt, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			isEncoded := false
			if encrypt != "plaintext" {
				isEncoded = true
			}

			for key, val := range data {
				encoded := string(val)
				if isEncoded {
					encoded, err = encryptor.Encrypt(string(val), appName)
					assert.NoError(t, err)
				}

				ss.ds[appName][key] = encoded
			}
		}

		for _, appName := range appNames {
			for _, scope := range scopes {
				entries, err := ss.ListEntries(appName, scope)
				assert.NoError(t, err)

				assert.Equal(t, len(data), len(entries))
			}
		}
	}
}

func TestDeleteEntry(t *testing.T) {
	for encrypt, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			isEncoded := false
			if encrypt != "plaintext" {
				isEncoded = true
			}

			for key, val := range data {
				encoded := string(val)
				if isEncoded {
					encoded, err = encryptor.Encrypt(string(val), appName)
					assert.NoError(t, err)
				}

				ss.ds[appName][key] = encoded
			}
		}

		for _, appName := range appNames {
			err := ss.DeleteEntry(appName, "secret", "key1")
			assert.NoError(t, err)

			for key := range ss.ds[appName] {
				assert.NotEqual(t, "key1", key)
			}
		}
	}
}

func TestListAppNames(t *testing.T) {
	for _, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		ss.client = NewMockClient(secrets["finex-public"], secrets["finex-private"], secrets["finex-secret"], secrets["storage-public"], secrets["storage-private"], secrets["storage-secret"])

		apps, err := ss.ListAppNames()
		assert.NoError(t, err)

		for _, appName := range appNames {
			found := false
			for i := range apps {
				if apps[i] == appName {
					found = true
					break
				}
			}
			assert.True(t, found)
		}
	}
}

func TestGetCurrentVersion(t *testing.T) {
	for _, encryptor := range encryptors {
		ss, err := NewService(deploymentID, client, encryptor)
		assert.NoError(t, err)

		if ss.ds == nil {
			ss.ds = make(map[string]map[string]interface{})
		}

		for _, appName := range appNames {
			if ss.ds[appName] == nil {
				ss.ds[appName] = make(map[string]interface{})
			}

			ss.ds[appName]["version"] = int64(3)
		}

		for _, appName := range appNames {
			for _, scope := range scopes {
				version, err := ss.GetCurrentVersion(appName, scope)
				assert.NoError(t, err)

				assert.Equal(t, int64(3), version)
			}
		}
	}
}
