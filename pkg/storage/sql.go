package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/openware/kaigara/pkg/encryptor/types"
	"github.com/openware/pkg/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SqlService contains a gorm DB client and a container for data loaded from DB into memory
type SqlService struct {
	db           *gorm.DB
	deploymentID string
	ds           map[string]map[string]map[string]interface{}
	encryptor    types.Encryptor
}

// Data represents per-scope data(configs/secrets) which consists of a JSON field
type SqlModel struct {
	gorm.Model
	AppName string
	Scope   string
	Value   datatypes.JSON
	Version int64
}

func NewSqlService(deploymentID string, conf *database.Config, encryptor types.Encryptor, logLevel int) (*SqlService, error) {
	conf.Name = "kaigara_" + deploymentID
	if err := ensureDatabaseExists(conf); err != nil {
		return nil, err
	}

	db, err := database.Connect(conf)
	if err != nil {
		return nil, err
	}
	db.Logger = logger.Default.LogMode(logger.LogLevel(logLevel))

	err = db.AutoMigrate(&SqlModel{})
	if err != nil {
		return nil, fmt.Errorf("SQL auto-migration failed: %s", err)
	}

	return &SqlService{
		db:           db,
		deploymentID: deploymentID,
		encryptor:    encryptor,
	}, nil
}

func ensureDatabaseExists(conf *database.Config) error {
	switch conf.Driver {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local",
			conf.User, conf.Pass, conf.Host, conf.Port,
		)
		conn, err := sql.Open(conf.Driver, dsn)
		if err != nil {
			return err
		}
		defer conn.Close()
		if _, err = conn.Exec("CREATE DATABASE IF NOT EXISTS " + conf.Name); err != nil {
			return err
		}
	case "postgres":
		dsn := fmt.Sprintf(
			"user=%s password=%s host=%s port=%s sslmode=disable",
			conf.User, conf.Pass, conf.Host, conf.Port,
		)
		conn, err := sql.Open(conf.Driver, dsn)
		if err != nil {
			return err
		}
		defer conn.Close()
		if res, err := conn.Exec(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", conf.Name)); err != nil {
			return err
		} else if rows, err := res.RowsAffected(); err != nil {
			return err
		} else if rows > 0 {
			return nil
		}
		if _, err = conn.Exec("CREATE DATABASE " + conf.Name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", conf.Driver)
	}

	return nil
}

func (ss *SqlService) Read(appName, scope string) error {
	var data SqlModel
	res := ss.db.First(&data, "app_name = ? AND scope = ?", appName, scope)

	if ss.ds == nil {
		ss.ds = make(map[string]map[string]map[string]interface{})
	}
	if ss.ds[appName] == nil {
		ss.ds[appName] = make(map[string]map[string]interface{})
	}
	if ss.ds[appName][scope] == nil {
		ss.ds[appName][scope] = make(map[string]interface{})
	}

	val := make(map[string]interface{})
	val["version"] = int64(0)

	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)
	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed reading from the DB: %s", res.Error)
	} else if !isNotFound {
		err := json.Unmarshal([]byte(data.Value), &val)
		if err != nil {
			return fmt.Errorf("JSON unmarshalling failed: %s", err)
		}

		val["version"] = data.Version
	}

	ss.ds[appName][scope] = val

	return nil
}

func (ss *SqlService) Write(appName, scope string) error {
	var ver int64
	if _, ok := ss.ds[appName][scope]["version"]; !ok {
		ss.ds[appName][scope]["version"] = 0
	} else if ver, ok = ss.ds[appName][scope]["version"].(int64); !ok {
		return fmt.Errorf("failed to get %s.%s.version: invalid value: %+v", appName, scope, ss.ds[appName][scope]["version"])
	}

	val := ss.ds[appName][scope]
	data := &SqlModel{
		AppName: appName,
		Scope:   scope,
		Version: int64(ver),
	}

	var old SqlModel
	res := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&old)
	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)
	isCreate := false

	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed to check for an existing value in the DB: %s", res.Error)
	} else if isNotFound {
		isCreate = true
	} else {
		data.Version = old.Version + 1
		val["version"] = data.Version
	}

	v, err := json.Marshal(val)
	if err != nil {
		return err
	}
	data.Value = v

	if isCreate {
		res = ss.db.Create(data)
		if res.Error != nil {
			return fmt.Errorf("initial DB record creation failed: %s", res.Error)
		}
	} else {
		err := res.Updates(data).Error
		if err != nil {
			return fmt.Errorf("existing DB record update failed: %s", err)
		}
	}

	return nil
}

func (ss *SqlService) ListEntries(appName, scope string) ([]string, error) {
	val, ok := ss.ds[appName][scope]
	if !ok {
		return []string{}, nil
	}

	res := make([]string, len(val))
	for k := range val {
		res = append(res, k)
	}

	return res, nil
}

func (ss *SqlService) SetEntry(appName, scope, name string, value interface{}) error {
	if scope == "secret" && name != "version" {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("sqlStore.SetSecret: %s is not a string", name)
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

func (ss *SqlService) SetEntries(appName string, scope string, values map[string]interface{}) error {
	for k, v := range values {
		err := ss.SetEntry(appName, scope, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ss *SqlService) GetEntry(appName, scope, name string) (interface{}, error) {
	// Since secret scope only supports strings, return a decrypted string
	scopeSecrets, ok := ss.ds[appName][scope]
	if !ok {
		return nil, fmt.Errorf("sqlStore.GetEntry: %s scope is not loaded", scope)
	}
	if scope == "secret" && name != "version" {
		rawValue, ok := scopeSecrets[name]
		if !ok {
			return nil, nil
		}

		str, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("sqlStore.GetEntry: %s is not a string", name)
		}

		decrypted, err := ss.encryptor.Decrypt(str, appName)
		if err != nil {
			return nil, err
		}

		return decrypted, nil
	}

	return ss.ds[appName][scope][name], nil
}

func (ss *SqlService) GetEntries(appName string, scope string) (map[string]interface{}, error) {
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

func (ss *SqlService) DeleteEntry(appName, scope, name string) error {
	delete(ss.ds[appName][scope], name)

	return nil
}

func (ss *SqlService) ListAppNames() ([]string, error) {
	var appNames []string
	tx := ss.db.Model(&SqlModel{}).Distinct().Pluck("app_name", &appNames)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return appNames, nil
}

func (ss *SqlService) GetCurrentVersion(appName, scope string) (int64, error) {
	if ss.ds[appName][scope] == nil {
		return 0, fmt.Errorf("failed to get %s.%s.version: scope is not loaded", appName, scope)
	}

	res, ok := ss.ds[appName][scope]["version"].(int64)
	if !ok {
		return 0, fmt.Errorf("failed to get %s.%s.version: type assertion to int64 failed, actual value: %v", appName, scope, res)
	}

	return res, nil
}

func (ss *SqlService) GetLatestVersion(appName, scope string) (int64, error) {
	var data SqlModel
	req := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&data)

	isNotFound := errors.Is(req.Error, gorm.ErrRecordNotFound)

	if req.Error != nil && !isNotFound {
		return 0, fmt.Errorf("failed to check for an existing value in the DB: %s", req.Error)
	} else if isNotFound {
		if ver, err := ss.GetCurrentVersion(appName, scope); err == nil {
			return ver, nil
		} else {
			return 0, nil
		}
	} else {
		return data.Version, nil
	}
}
