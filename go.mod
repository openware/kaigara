module github.com/openware/kaigara

go 1.17

// replace github.com/openware/kaigara/pkg/encryptor => ./pkg/encryptor

// replace github.com/openware/kaigara/pkg/sql => ./pkg/sql

// replace github.com/openware/kaigara/pkg/vault => ./pkg/vault

require (
	github.com/go-redis/redis/v7 v7.2.0
	github.com/openware/kaigara/pkg/encryptor v0.0.0-20220428165818-6271445f8750
	github.com/openware/kaigara/pkg/sql v0.0.0-20220512125342-51f54b8d8897
	github.com/openware/kaigara/pkg/vault v0.0.0-20220428165818-6271445f8750
	github.com/openware/pkg v0.0.0-20220225074124-ddad5f429a07
	github.com/stretchr/testify v1.7.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/gorm v1.22.5
)
