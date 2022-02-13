package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// StorageService contains a gorm DB client and a container for data loaded from DB into memory
type StorageService struct {
	db           *gorm.DB
	deploymentID string
	Datasets     map[string]map[string]Data
}

type Data struct {
	gorm.Model
	Scope   string
	Value   string
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
		return fmt.Errorf("StorageService: Read: %s", res.Error)
	}

	ss.Datasets[appName][scope] = data

	return nil
}
