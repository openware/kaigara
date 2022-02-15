package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/types"
	"github.com/openware/pkg/database"
	"github.com/stretchr/testify/assert"
)

var deploymentID string
var appNames []string
var sqlConfig *database.Config

var stores = make(map[string]types.Storage)
var envs = map[string]map[string]map[string]string{
	"finex": {
		"public": {},
		"private": {
			"finex_database_driver":        "mysql",
			"finex_database_host":          "mysql-v4.core",
			"finex_database_port":          "3306",
			"finex_gotrue_pub_key":         "AeZ0tie1thae6lee9red",
			"finex_influx_host":            "influxdb-0.core,influxdb-1.core,influxdb-2.core",
			"finex_influx_port":            "8086",
			"finex_influx_scheme":          "http",
			"finex_key":                    "ga8eiphahz0Ooph8Eih0",
			"finex_log_level":              "debug",
			"finex_mode":                   "prod",
			"finex_vault_addr":             "http://vault.core:8200",
			"finex_vault_app":              "opendax_uat",
			"finex_vault_contract_address": "Xohthoh5ohl7eepihuul",
			"require_quote_amount_roles":   "sa_maker",
		},
		"secret": {
			"finex_database_name":            "finex_opendax_uat",
			"finex_database_password":        "fuc2KeGio6paekiefahn",
			"finex_database_username":        "finex_opendax_uat",
			"finex_deployment_id":            "opendax_uat",
			"finex_influx_database":          "finex_opendax_uat",
			"finex_influx_password":          "zie8uPhe2aebae9viroh",
			"finex_influx_username":          "opendax",
			"finex_license_key":              "eiJohdo9eish3Cooshus",
			"finex_vault_broker_private_key": "ujahDoo1ohmie7taipox",
			"finex_vault_rpc_url":            "wss://rinkeby.infura.io/ws/v3/Uzeep6eiGoozui7ohsh9",
		},
	},
	"frontdex": {
		"public": {},
		"private": {
			"next_public_chain_id":         "4",
			"next_public_custody_contract": "Feibie7ooCachie3eePh",
		},
		"secret": {
			"next_public_infura_id": "bie9niNgoohadoorai5a",
		},
	},
	"global": {
		"public":  {},
		"private": {},
		"secret": {
			"database_host": "0.0.0.0",
			"database_port": "3306",
			"postgres_host": "postgresql.core",
			"postgres_port": "5432",
		},
	},
	"gotrue": {
		"public": {},
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
			"gotrue_operator_token":               "Jo8eengeiz4poocohtei",
			"gotrue_smtp_admin_email":             "riashaid7di1HooBoo3z",
			"gotrue_smtp_host":                    "smtp.mailgun.org",
			"gotrue_smtp_pass":                    "quoiN0Iegoote7hoik0V",
			"gotrue_smtp_port":                    "25",
			"gotrue_smtp_user":                    "QueeWai1Dailef6ies3y",
		},
		"secret": {
			"database_url":             "postgres://gotrue_odax_yellow_com:thaiDooJohw9baechaiH@postgresql.core:5432/opendax_odax_yellow_com",
			"gotrue_database_host":     "postgresql.core",
			"gotrue_database_name":     "opendax_odax_yellow_com",
			"gotrue_database_password": "eiyehiaFei0eing4Caiy",
			"gotrue_database_username": "gotrue_odax_yellow_com",
			"gotrue_disable_signup":    "false",
			"gotrue_jwt_secret":        "ahteebieHo1Eequ2Eeth",
		},
	},
	"postgrest": {
		"public":  {},
		"private": {},
		"secret": {
			"pgrst_db_host":     "postgresql.core",
			"pgrst_db_name":     "opendax_odax_yellow_com",
			"pgrst_db_pass":     "iey2Mei1aib5ioz0Kai3",
			"pgrst_db_port":     "5432",
			"pgrst_db_uri":      "postgres://postgrest_odax_yellow_com:ricuPheechau8feezahb@postgresql.core:5432/opendax_odax_yellow_com",
			"pgrst_db_username": "postgrest_odax_yellow_com",
		},
	},
	"realtime": {
		"public":  {},
		"private": {},
		"secret": {
			"realtime_db_host":     "postgresql.core",
			"realtime_db_name":     "opendax_odax_yellow_com",
			"realtime_db_pass":     "bahchiePaeh0eeDuoW2i",
			"realtime_db_username": "realtime_odax_yellow_com",
		},
	},
	"storage": {
		"public":  {},
		"private": {},
		"secret": {
			"storage_db_host":     "postgresql.core",
			"storage_db_name":     "opendax_odax_yellow_com",
			"storage_db_pass":     "esaishahSe7Amu7Iefi4",
			"storage_db_uri":      "postgres://storage_odax_yellow_com:phocie0Ma8ooFalei6Ah@postgresql.core:5432/opendax_odax_yellow_com",
			"storage_db_username": "storage_odax_yellow_com",
		},
	},
}
var vars = []string{
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

func getStorage(driver string, cfg *config.KaigaraConfig, db *database.Config) types.Storage {
	store := stores[driver]
	if store == nil {
		store = config.GetStorageService(cfg, db)

		for name, app := range envs {
			for scope, elem := range app {
				store.Read(name, scope)
				secrets, err := store.GetEntries(name, scope)
				if err != nil {
					panic(err)
				}
				isSave := false
				for key, val := range elem {
					if _, ok := secrets[key]; !ok {
						isSave = true
						store.SetEntry(name, scope, key, val)
					}
				}
				if isSave {
					store.Write(name, scope)
				}
			}
		}
	}
	return store
}

func TestMain(m *testing.M) {
	deploymentID = "opendax_uat"
	appNames = []string{"finex", "gotrue", "postgrest", "realtime", "storage"}
	sqlConfig = &database.Config{
		Driver: "mysql",
		Host:   "0.0.0.0",
		Port:   "3306",
		Name:   "kaigara_dev",
		User:   "root",
		Pass:   "",
		Pool:   1,
	}

	// exec test and this returns an exit code to pass to os
	code := m.Run()

	os.Exit(code)
}

func TestKaidumpListAppNames(t *testing.T) {
	db := sqlConfig
	cfg := &config.KaigaraConfig{
		SecretStore:  "sql",
		VaultAddr:    os.Getenv("KAIGARA_VAULT_ADDR"),
		VaultToken:   os.Getenv("KAIGARA_VAULT_TOKEN"),
		DeploymentID: deploymentID,
		Scopes:       "private,secret",
		AppNames:     "finex,gotrue,postgrest,realtime,storage",
	}

	cnf = cfg
	sqlCnf = db
	store := getStorage("sql", cfg, db)
	b := kaidumpRun(store)
	assert.NotNil(t, b)
	fmt.Print(b.String())
}
