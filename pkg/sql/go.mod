module github.com/openware/kaigara/pkg/storage/sql

go 1.14

replace github.com/openware/kaigara/pkg/encryptor => ../encryptor

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/lib/pq v1.10.2
	github.com/openware/kaigara/pkg/encryptor v0.0.0-00010101000000-000000000000
	github.com/openware/pkg v0.0.0-20220225074124-ddad5f429a07
	github.com/stretchr/testify v1.7.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/datatypes v1.0.6
	gorm.io/gorm v1.23.2
)
