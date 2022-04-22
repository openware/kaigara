module github.com/openware/kaigara/pkg/storage/vault

go 1.14

replace github.com/openware/kaigara/pkg/encryptor => ../../encryptor

require (
	github.com/hashicorp/vault/api v1.3.1
	github.com/iancoleman/strcase v0.2.0
	github.com/stretchr/testify v1.7.1
)
