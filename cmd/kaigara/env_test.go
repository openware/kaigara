package main

import (
	"os"
	"testing"

	"github.com/openware/kaigara/pkg/vault"
)

var store *vault.Service
var envs = map[string]map[string]map[string]string{
	"finex": {
		"private": {
			"finex_database_driver":        "mysql",
			"finex_database_host":          "mysql-v4.core",
			"finex_database_port":          "3306",
			"finex_gotrue_pub_key":         "changeme",
			"finex_influx_host":            "influxdb-0.core,influxdb-1.core,influxdb-2.core",
			"finex_influx_port":            "8086",
			"finex_influx_scheme":          "http",
			"finex_key":                    "changeme",
			"finex_log_level":              "debug",
			"finex_mode":                   "prod",
			"finex_vault_addr":             "http://vault.core:8200",
			"finex_vault_app":              "opendax_uat",
			"finex_vault_contract_address": "changeme",
			"require_quote_amount_roles":   "sa_maker",
		},
		"secret": {
			"finex_database_name":            "finex_opendax_uat",
			"finex_database_password":        "changeme",
			"finex_database_username":        "finex_opendax_uat",
			"finex_deployment_id":            "opendax_uat",
			"finex_influx_database":          "finex_opendax_uat",
			"finex_influx_password":          "changeme",
			"finex_influx_username":          "opendax",
			"finex_license_key":              "changeme",
			"finex_vault_broker_private_key": "changeme",
			"finex_vault_rpc_url":            "wss://rinkeby.infura.io/ws/v3/changeme",
		},
	},
	"frontdex": {
		"private": {
			"next_public_chain_id":         "4",
			"next_public_custody_contract": "changeme",
		},
		"secret": {
			"next_public_infura_id": "changeme",
		},
	},
	"global": {
		"secret": {
			"database_host": "mysql-v4.core",
			"database_port": "3306",
			"postgres_host": "postgresql.core",
			"postgres_port": "5432",
		},
	},
	"gotrue": {
		"private": {
			"db_namespace":                        "auth",
			"gotrue_api_host":                     "0.0.0.0",
			"gotrue_cookie_duration":              "3600",
			"gotrue_db_driver":                    "postgres",
			"gotrue_jwt_algorithm":                "RS256",
			"gotrue_jwt_default_group_name":       "user",
			"gotrue_jwt_exp":                      "3600",
			"gotrue_mailer_autoconfirm":           "true",
			"gotrue_mailer_urlpaths_confirmation": "/verify",
			"gotrue_mailer_urlpaths_invite":       "/verify",
			"gotrue_mailer_urlpaths_recovery":     "/verify",
			"gotrue_operator_token":               "changeme",
			"gotrue_smtp_admin_email":             "changeme",
			"gotrue_smtp_host":                    "smtp.mailgun.org",
			"gotrue_smtp_pass":                    "changeme",
			"gotrue_smtp_port":                    "25",
			"gotrue_smtp_user":                    "changeme",
		},
		"secret": {
			"database_url":             "postgres://gotrue_odax_yellow_com:changeme@postgresql.core:5432/opendax_odax_yellow_com",
			"gotrue_database_host":     "postgresql.core",
			"gotrue_database_name":     "opendax_odax_yellow_com",
			"gotrue_database_password": "changeme",
			"gotrue_database_username": "gotrue_odax_yellow_com",
			"gotrue_disable_signup":    "false",
			"gotrue_jwt_secret":        "changeme",
		},
	},
	"postgrest": {
		"secret": {
			"pgrst_db_host":     "postgresql.core",
			"pgrst_db_name":     "opendax_odax_yellow_com",
			"pgrst_db_pass":     "changeme",
			"pgrst_db_port":     "5432",
			"pgrst_db_uri":      "postgres://postgrest_odax_yellow_com:changeme@postgresql.core:5432/opendax_odax_yellow_com",
			"pgrst_db_username": "postgrest_odax_yellow_com",
		},
	},
	"realtime": {
		"secret": {
			"realtime_db_host":     "postgresql.core",
			"realtime_db_name":     "opendax_odax_yellow_com",
			"realtime_db_pass":     "changeme",
			"realtime_db_username": "realtime_odax_yellow_com",
		},
	},
	"storage": {
		"secret": {
			"storage_db_host":     "postgresql.core",
			"storage_db_name":     "opendax_odax_yellow_com",
			"storage_db_pass":     "changeme",
			"storage_db_uri":      "postgres://storage_odax_yellow_com:changeme@postgresql.core:5432/opendax_odax_yellow_com",
			"storage_db_username": "storage_odax_yellow_com",
		},
	},
}

func getSecretStore() *vault.Service {
	if store == nil {
		store = getVaultService()

		for name, app := range envs {
			for scope, elem := range app {
				store.LoadSecrets(name, scope)
				secrets, err := store.GetSecrets(name, scope)
				if err != nil {
					panic(err)
				}
				isSave := false
				for key, val := range elem {
					if _, ok := secrets[key]; !ok {
						isSave = true
						store.SetSecret(name, key, val, scope)
					}
				}
				if isSave {
					store.SaveSecrets(name, scope)
				}
			}
		}
	}
	return store
}

// TestMain will exec each test, one by one
func TestMain(m *testing.M) {
	cnf.VaultAddr = os.Getenv("KAIGARA_VAULT_ADDR")
	cnf.VaultToken = os.Getenv("KAIGARA_VAULT_TOKEN")
	cnf.DeploymentID = "opendax_uat"
	cnf.Scopes = "private,secret"
	cnf.AppNames = "finex,gotrue,postgrest,realtime,storage"

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	// secretStore := getSecretStore()
	// for name, app := range envs {
	// 	for scope, elem := range app {
	// 		secretStore.LoadSecrets(name, scope)
	// 		for key, _ := range elem {
	// 			secretStore.DeleteSecret(name, key, scope)
	// 		}
	// 	}
	// }

	os.Exit(code)
}

func TestKaigaraPrintenv(t *testing.T) {
	secretStore := getSecretStore()

	ls := initLogStream()
	vars := []string{
		"FINEX_DATABASE_USERNAME",
		"FINEX_DATABASE_PASSWORD",
		"FINEX_DATABASE_NAME",
		"FINEX_DATABASE_HOST",
		"FINEX_INFLUX_USERNAME",
		"FINEX_INFLUX_PASSWORD",
		"FINEX_INFLUX_HOST",
		"GOTRUE_DATABASE_USERNAME",
		"GOTRUE_DATABASE_PASSWORD",
		"GOTRUE_DATABASE_NAME",
		"GOTRUE_DATABASE_HOST",
		"PGRST_DB_USERNAME",
		"PGRST_DB_PASS",
		"PGRST_DB_NAME",
		"PGRST_DB_HOST",
		"REALTIME_DB_USERNAME",
		"REALTIME_DB_PASS",
		"REALTIME_DB_NAME",
		"REALTIME_DB_HOST",
	}
	for _, v := range vars {
		kaigaraRun(ls, secretStore, "printenv", []string{v})
	}
}
