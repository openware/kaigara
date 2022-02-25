module github.com/openware/kaigara/pkg/storage/sql

go 1.14

replace github.com/openware/kaigara/pkg/encryptor => ../../../pkg/encryptor

require (
	github.com/openware/kaigara/pkg/encryptor v0.0.0
	github.com/openware/pkg v0.0.0-20220225074124-ddad5f429a07
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/datatypes v1.0.5
	gorm.io/driver/mysql v1.2.3 // indirect
	gorm.io/driver/sqlite v1.2.6 // indirect
	gorm.io/driver/sqlserver v1.2.1 // indirect
	gorm.io/gorm v1.22.5
)
