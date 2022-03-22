#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=changeme

FILE=./bin/kaigara
if [ ! -f "$FILE" ]; then
  CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara
fi

docker-compose -f ./etc/backend.yml down -v
docker-compose -f ./etc/backend.yml up -d
vault secrets enable transit

db_ready() {
  mysqladmin ping --host=$DATABASE_HOST --user=$DATABASE_USER --password=$DATABASE_PASS >/dev/null 2>&1
}

export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
export KAIGARA_DEPLOYMENT_ID=opendax_uat
export KAIGARA_SCOPES=public,private,secret
export KAIGARA_APP_NAME=finex,frontdex,gotrue,postgrest,realtime,storage
export KAIGARA_ENCRYPTOR_AES_KEY=changemechangemechangeme

export DATABASE_HOST="0.0.0.0"
export DATABASE_NAME=kaigara_dev
export DATABASE_POOL=1

ASSERTS=("FINEX_DATABASE_USERNAME:finex_opendax_uat"
  "FINEX_DATABASE_PASSWORD:fuc2KeGio6paekiefahn"
  "FINEX_DATABASE_NAME:finex_opendax_uat"
  "FINEX_DATABASE_HOST:mysql-v4.core"
  "FINEX_INFLUX_USERNAME:opendax"
  "FINEX_INFLUX_PASSWORD:zie8uPhe2aebae9viroh"
  "FINEX_INFLUX_HOST:influxdb-0.core,influxdb-1.core,influxdb-2.core"
  "GOTRUE_DATABASE_USERNAME:gotrue_odax_yellow_com"
  "GOTRUE_DATABASE_PASSWORD:eiyehiaFei0eing4Caiy"
  "GOTRUE_DATABASE_NAME:opendax_odax_yellow_com"
  "GOTRUE_DATABASE_HOST:postgresql.core"
  "PGRST_DB_USERNAME:postgrest_odax_yellow_com"
  "PGRST_DB_PASS:iey2Mei1aib5ioz0Kai3"
  "PGRST_DB_NAME:opendax_odax_yellow_com"
  "PGRST_DB_HOST:postgresql.core"
  "REALTIME_DB_USERNAME:realtime_odax_yellow_com"
  "REALTIME_DB_PASS:bahchiePaeh0eeDuoW2i"
  "REALTIME_DB_NAME:opendax_odax_yellow_com"
  "REALTIME_DB_HOST:postgresql.core")

DELETES=("FINEX_DATABASE_HOST"
  "FINEX_DATABASE_PASSWORD")

SECRET_STORES=("vault:vault"
  "sql:mysql"
  "sql:postgres")

while !(db_ready); do
  sleep 3
  echo "waiting for db ..."
done

for store in "${SECRET_STORES[@]}"; do
  ss="${store%%:*}"
  driver="${store##*:}"

  export KAIGARA_STORAGE_DRIVER="${ss}"

  export DATABASE_DRIVER="${driver}"
  if [ "${driver}" == "mysql" ]; then
    export DATABASE_PORT=3306
    export DATABASE_USER=root
    export DATABASE_PASS=
  fi

  if [ "${driver}" == "postgres" ]; then
    export DATABASE_PORT=5432
    export DATABASE_USER=postgres
    export DATABASE_PASS=changeme
  fi

  env | grep KAIGARA
  env | grep DATABASE

  ./bin/kaisave -filepath ./scripts/odax_values.yml

  for assert in "${ASSERTS[@]}"; do
    KEY="${assert%%:*}"
    VALUE="${assert##*:}"

    out=$(./bin/kaienv $KEY)
    if [ "$out" == "$VALUE" ]; then
      echo "ENV: { Key: ${KEY}, Out: ${out} }"
    else
      echo "ERROR: { Key: ${KEY}, Out: ${out} }"
      exit 1
    fi
  done

  TMP_FILE=./tmp/odax_tmp.yml
  ./bin/kaidump -a $TMP_FILE
  if [ ! -f "$TMP_FILE" ]; then
    exit 1
  fi
  cat $TMP_FILE

  ./bin/kaidel -a finex -s private -k finex_database_host
  ./bin/kaidel -a finex -s secret -k finex_database_password

  for assert in "${DELETES[@]}"; do
    KEY="${assert}"

    if ./bin/kaienv $KEY; then
      exit 1
    fi
  done
done
