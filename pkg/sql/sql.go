package sql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openware/kaigara/pkg/encryptor/types"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// Service contains a gorm DB client and a container for data loaded from DB into memory
type Service struct {
	db           *gorm.DB
	deploymentID string
	ds           map[string]map[string]map[string]interface{}
	encryptor    types.Encryptor
}

// Data represents per-scope data(configs/secrets) which consists of a JSON field
type Data struct {
	gorm.Model
	AppName string
	Scope   string
	Value   datatypes.JSON
	Version int64
}

type DatabaseConfig struct {
	Driver string `yaml:"driver" env:"KAIGARA_DATABASE_DRIVER" env-description:"Database driver"`
	Host   string `yaml:"host" env:"KAIGARA_DATABASE_HOST" env-description:"Database host"`
	Port   string `yaml:"port" env:"KAIGARA_DATABASE_PORT" env-description:"Database port"`
	Name   string `yaml:"name" env:"KAIGARA_DATABASE_NAME" env-description:"Database name"`
	User   string `yaml:"user" env:"KAIGARA_DATABASE_USER" env-description:"Database user"`
	Pass   string `env:"KAIGARA_DATABASE_PASS" env-description:"Database user password"`
	Pool   int    `yaml:"pool" env:"KAIGARA_DATABASE_POOL" env-description:"Database pool size" env-default:"0"`
}

func NewService(deploymentID string, conf *DatabaseConfig, encryptor types.Encryptor, logLevel int) (*Service, error) {
	conf.Name = "kaigara_" + deploymentID
	if err := EnsureDatabaseExists(conf); err != nil {
		return nil, err
	}
	db, err := Connect(conf)
	if err != nil {
		return nil, err
	}
	db.Logger = logger.Default.LogMode(logger.LogLevel(logLevel))

	if logLevel > 1 {
		db = db.Debug()
	}

	err = db.AutoMigrate(&Data{})
	if err != nil {
		return nil, fmt.Errorf("SQL auto-migration failed: %s", err)
	}

	return &Service{
		db:           db,
		deploymentID: deploymentID,
		encryptor:    encryptor,
	}, nil
}

func EnsureDatabaseExists(cnf *DatabaseConfig) error {
	switch cnf.Driver {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local",
			cnf.User, cnf.Pass, cnf.Host, cnf.Port,
		)
		conn, err := sql.Open(cnf.Driver, dsn)
		if err != nil {
			return err
		}
		defer conn.Close()
		if _, err = conn.Exec("CREATE DATABASE IF NOT EXISTS " + cnf.Name); err != nil {
			return err
		}
	case "postgres":
		dsn := fmt.Sprintf(
			"user=%s password=%s host=%s port=%s dbname=postgres sslmode=disable",
			cnf.User, cnf.Pass, cnf.Host, cnf.Port,
		)
		conn, err := sql.Open(cnf.Driver, dsn)
		if err != nil {
			return err
		}
		defer conn.Close()
		if res, err := conn.Exec(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", cnf.Name)); err != nil {
			return err
		} else if rows, err := res.RowsAffected(); err != nil {
			return err
		} else if rows > 0 {
			return nil
		}
		if _, err = conn.Exec("CREATE DATABASE " + cnf.Name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", cnf.Driver)
	}

	return nil
}

func Connect(conf *DatabaseConfig) (*gorm.DB, error) {
	var err error
	var dial gorm.Dialector
	var dsn string

	switch conf.Driver {
	case "mysql":
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			conf.User, conf.Pass, conf.Host, conf.Port, conf.Name,
		)
		dial = mysql.Open(dsn)

	case "postgres":
		dsn := fmt.Sprintf(
			"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
			conf.User, conf.Pass, conf.Host, conf.Port, conf.Name,
		)
		dial = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported DATABASE_DRIVER: %s", conf.Driver)
	}

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// FIXME: Decide whether to keep block
	// The default value is 0, we should check whether we need to set it or not
	// Currently, sql.SetMaxOpenConns with conf.Pool set to 1 causes the code to stall infinitely
	if conf.Pool != 0 {
		sql, err := db.DB()
		if err != nil {
			return nil, err
		}
		sql.SetMaxOpenConns(conf.Pool)
	}

	return db, nil
}

func (ss *Service) Read(appName, scope string) error {
	var data Data
	res := ss.db.First(&data, "app_name = ? AND scope = ?", appName, scope)

	if ss.ds == nil {
		ss.ds = make(map[string]map[string]map[string]interface{})
	}
	if ss.ds[appName] == nil {
		ss.ds[appName] = make(map[string]map[string]interface{})
	}

	val := make(map[string]interface{})
	val["version"] = int64(0)

	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)
	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed reading from the DB: %s", res.Error)
	} else if !isNotFound {
		if err := json.Unmarshal([]byte(data.Value), &val); err != nil {
			return fmt.Errorf("JSON unmarshalling failed: %s", err)
		}

		val["version"] = data.Version
	}

	ss.ds[appName][scope] = val

	return nil
}

func (ss *Service) Write(appName, scope string) error {
	val, ok := ss.ds[appName][scope]
	if !ok {
		return fmt.Errorf("scope '%s' in '%s' app is: %v", scope, appName, val)
	}

	data := &Data{
		AppName: appName,
		Scope:   scope,
	}

	var old Data
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

func (ss *Service) SetEntries(appName string, scope string, values map[string]interface{}) error {
	for k, v := range values {
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

func (ss *Service) GetEntries(appName string, scope string) (map[string]interface{}, error) {
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

func (ss *Service) DeleteEntry(appName, scope, name string) error {
	delete(ss.ds[appName][scope], name)

	return nil
}

func (ss *Service) ListAppNames() ([]string, error) {
	var appNames []string
	tx := ss.db.Model(&Data{}).Distinct().Pluck("app_name", &appNames)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return appNames, nil
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
	var data Data
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
