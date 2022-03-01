module github.com/openware/kaigara

go 1.14

replace github.com/openware/kaigara/pkg/encryptor => ./pkg/encryptor

replace github.com/openware/kaigara/pkg/storage/sql => ./pkg/storage/sql

replace github.com/openware/kaigara/pkg/storage/vault => ./pkg/storage/vault

require (
	github.com/go-redis/redis/v7 v7.2.0
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/mattn/go-sqlite3 v1.14.11 // indirect
	github.com/openware/kaigara/pkg/encryptor v0.0.0-20220225091359-d368f0dfe8db
	github.com/openware/kaigara/pkg/storage/sql v0.0.0-20220301034206-c8eea45f3512
	github.com/openware/kaigara/pkg/storage/vault v0.0.0-20220301034206-c8eea45f3512
	github.com/openware/pkg v0.0.0-20220225074124-ddad5f429a07
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20211215165025-cf75a172585e // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/gorm v1.22.5
)
