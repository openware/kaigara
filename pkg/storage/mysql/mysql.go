package mysql

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// StorageService contains a gorm DB client and a container for data loaded from DB into memory
type StorageService struct {
	db           *gorm.DB
	deploymentID string
	Datastore    map[string]map[string]map[string]interface{}
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &StorageService{
		db:           db,
		deploymentID: deploymentID,
	}, nil
}

func (ss *StorageService) Read(appName, scope string) error {
	var data Data
	res := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&data)
	if res.Error != nil {
		return fmt.Errorf("read from DB failed: %s", res.Error)
	}

	val := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Value), &val)
	if err != nil {
		return fmt.Errorf("JSON unmarshalling failed: %s", err)
	}

	ss.Datastore[appName][scope] = val

	return nil
}

func (ss *StorageService) Write(appName, scope string) error {
	val := ss.Datastore[appName][scope]

	v, err := json.Marshal(val)
	if err != nil {
		return err
	}

	data := Data{
		AppName: appName,
		Scope:   scope,
		Value:   v,
	}

	res := ss.db.Where("app_name = ? AND scope = ?", appName, scope).First(&data)
	isNotFound := errors.Is(res.Error, gorm.ErrRecordNotFound)

	if res.Error != nil && !isNotFound {
		return fmt.Errorf("failed to check for an existing value in the DB: %s", res.Error)
	} else if isNotFound {
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

func (ss *StorageService) ListEntries(appName, scope string) ([]string, error) {
	val, ok := ss.Datastore[appName][scope]
	if !ok {
		return []string{}, nil
	}

	res := make([]string, len(val))
	for k := range val {
		res = append(res, k)
	}

	return res, nil
}
