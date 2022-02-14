package mysql

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// StorageService contains a gorm DB client and a container for data loaded from DB into memory
type StorageService struct {
	db           *gorm.DB
	deploymentID string
	ds           map[string]map[string]map[string]interface{}
}

// Data represents per-scope data(configs/secrets) which consists of a JSON field
type Data struct {
	gorm.Model
	AppName string
	Scope   string
	Value   datatypes.JSON
	Version uint
}

func NewStorageService(dsn, deploymentID string) (*StorageService, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Data{})
	if err != nil {
		return nil, fmt.Errorf("SQL auto-migration failed: %s", err)
	}

	return &StorageService{
		db:           db,
		deploymentID: deploymentID,
	}, nil
}

func (ss *StorageService) Read(appName, scope string) error {
	var data Data
	res := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&data)

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

	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)
	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed reading from the DB: %s", res.Error)
	} else if !isNotFound {
		fmt.Printf("INFO: reading %s.%s from DB\n", appName, scope)
		err := json.Unmarshal([]byte(data.Value), &val)
		if err != nil {
			return fmt.Errorf("JSON unmarshalling failed: %s", err)
		}
	}

	ss.ds[appName][scope] = val

	return nil
}

func (ss *StorageService) Write(appName, scope string) error {
	val := ss.ds[appName][scope]

	v, err := json.Marshal(val)
	if err != nil {
		return err
	}

	fresh := &Data{
		AppName: appName,
		Scope:   scope,
		Value:   v,
	}

	var old Data
	res := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&old)

	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)

	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed to check for an existing value in the DB: %s", res.Error)
	} else if isNotFound {
		res = ss.db.Create(fresh)
		if res.Error != nil {
			return fmt.Errorf("initial DB record creation failed: %s", res.Error)
		}
	} else {
		fresh.Version = old.Version + 1
		err := res.Updates(fresh).Error
		if err != nil {
			return fmt.Errorf("existing DB record update failed: %s", err)
		}
	}

	return nil
}

func (ss *StorageService) ListEntries(appName, scope string) ([]string, error) {
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

func (ss *StorageService) SetEntry(appName, scope, name string, value interface{}) error {
	fmt.Printf("SetEntry: before: scope %+v\n", ss.ds[appName][scope])
	ss.ds[appName][scope][name] = value
	fmt.Printf("SetEntry: before: scope %+v\n", ss.ds[appName][scope])
	return nil
}

func (ss *StorageService) SetEntries(appName string, scope string, values map[string]interface{}) error {
	for k, v := range values {
		ss.SetEntry(appName, scope, k, v)
	}
	return nil
}

func (ss *StorageService) GetEntry(appName, scope, name string) (interface{}, error) {
	fmt.Printf("GetEntry: before: scope %+v\n", ss.ds[appName][scope])
	if ss.ds[appName][scope] == nil {
		return nil, fmt.Errorf("failed to get %s.%s.%s: scope is not loaded", appName, scope, name)
	}
	return ss.ds[appName][scope][name], nil
}

func (ss *StorageService) GetEntries(appName string, scope string) (map[string]interface{}, error) {
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

func (ss *StorageService) DeleteEntry(appName, scope, name string) error {
	fmt.Printf("Deleting %s.%s.%s\n", appName, scope, name)
	delete(ss.ds[appName][scope], name)

	return nil
}
