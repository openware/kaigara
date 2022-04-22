module github.com/openware/kaigara

go 1.14

replace github.com/openware/kaigara/pkg/encryptor => ./pkg/encryptor

replace github.com/openware/kaigara/pkg/sql => ./pkg/sql

replace github.com/openware/kaigara/pkg/vault => ./pkg/vault

require (
	github.com/go-redis/redis/v7 v7.2.0
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/openware/kaigara/pkg/encryptor v0.0.0-00010101000000-000000000000
	github.com/openware/kaigara/pkg/sql v0.0.0-00010101000000-000000000000
	github.com/openware/kaigara/pkg/vault v0.0.0-00010101000000-000000000000
	github.com/openware/pkg v0.0.0-20220225074124-ddad5f429a07
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/stretchr/testify v1.7.1
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/gorm v1.23.2
)
