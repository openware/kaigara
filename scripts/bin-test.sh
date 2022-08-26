#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=changeme

GOOS=$(uname -s | tr A-Z a-z)
GOARCH=$(uname -m)
FILE=./bin/kai
if [ ! -f "$FILE" ]; then
  env GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -a -ldflags "-w -X main.Version=${KAIGARA_VERSION}" -o ${FILE} ./cmd/kaicli
fi

docker-compose -f ./etc/backend.yml down -v
docker-compose -f ./etc/backend.yml up -d
vault secrets enable transit

db_ready() {
  mysqladmin ping --host=$DATABASE_HOST --user=$DATABASE_USER --password=$DATABASE_PASS >/dev/null 2>&1
}

export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
export KAIGARA_DEPLOYMENT_ID=odax
export KAIGARA_SCOPES=public,private,secret
export KAIGARA_APP_NAME=finex,frontdex,gotrue,postgrest,realtime,storage
export KAIGARA_ENCRYPTOR=plaintext
export KAIGARA_ENCRYPTOR_AES_KEY=changemechangemechangeme
export KUBECONFIG=~/.kube/microk8s

export DATABASE_HOST="0.0.0.0"
export DATABASE_POOL=1

ASSERTS=("FINEX_DATABASE_USERNAME:finex_opendax_uat"
  "FINEX_DATABASE_PASSWORD:fuc2KeGio6paekiefahn"
  "FINEX_DATABASE_NAME:finex_opendax_uat"
  "FINEX_DATABASE_HOST:mysql-v4.core"
  "FINEX_INFLUX_USERNAME:opendax"
  "FINEX_INFLUX_PASSWORD:zie8uPhe2aebae9viroh"
  "FINEX_INFLUX_HOST:influxdb-0.core,influxdb-1.core,influxdb-2.core"
  "GOTRUE_DATABASE_USERNAME:gotrue_odax_example"
  "GOTRUE_DATABASE_PASSWORD:eiyehiaFei0eing4Caiy"
  "GOTRUE_DATABASE_NAME:opendax_odax_example"
  "GOTRUE_DATABASE_HOST:postgresql.core"
  "PGRST_DB_USERNAME:postgrest_odax_example"
  "PGRST_DB_PASS:iey2Mei1aib5ioz0Kai3"
  "PGRST_DB_NAME:opendax_odax_example"
  "PGRST_DB_HOST:postgresql.core"
  "REALTIME_DB_USERNAME:realtime_odax_example"
  "REALTIME_DB_PASS:bahchiePaeh0eeDuoW2i"
  "REALTIME_DB_NAME:opendax_odax_example"
  "REALTIME_DB_HOST:postgresql.core")

DELETES=("FINEX_DATABASE_HOST"
  "FINEX_DATABASE_PASSWORD")

SECRET_STORES=("vault:vault"
  "sql:mysql"
  "sql:postgres"
  "k8s:k8s")

for ss in "${SECRET_STORES[@]}"; do
  ss="${ss%%:*}"
  driver="${ss##*:}"

  export KAIGARA_STORAGE_DRIVER="${ss}"

  if [ "${ss}" == "sql" ]; then
    while !(db_ready); do
      sleep 3
      echo "waiting for db ..."
    done
  fi

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

  if [ "${ss}" == "k8s" ]; then
    export KAIGARA_SCOPES=secret
  fi

  env | grep KAIGARA
  env | grep DATABASE

  ./bin/kai save -f ./scripts/odax_values.yml

  for assert in "${ASSERTS[@]}"; do
    KEY="${assert%%:*}"
    VALUE="\"${assert##*:}\""

    out=$(./bin/kai env $KEY)
    if [ "$out" == "$VALUE" ]; then
      echo "ENV: { Key: ${KEY}, Expect: ${VALUE}, Actual: ${out} }"
    else
      echo "ERROR: { Key: ${KEY}, Expect: ${VALUE}, Actual: ${out} }"
      exit 1
    fi
  done

  TMP_FILE=./tmp/odax_tmp.yml
  ./bin/kai dump -o $TMP_FILE
  if [ ! -f "$TMP_FILE" ]; then
    exit 1
  fi
  cat $TMP_FILE

  ./bin/kai del finex.private.finex_database_host
  ./bin/kai del finex.secret.finex_database_password

  for assert in "${DELETES[@]}"; do
    KEY="${assert}"

    if ./bin/kai env $KEY; then
      exit 1
    fi
  done
done
