#!/bin/bash

export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=root-token
export KAIGARA_DEPLOYMENT_ID=opendax_uat
export KAIGARA_SCOPES=public,private,secret
export KAIGARA_APP_NAME=finex,gotrue,postgrest,realtime,storage

ROUND=$1

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

for assert in "${ASSERTS[@]}"; do
  KEY="${assert%%:*}"
  VALUE="${assert##*:}"

  out=$(./bin/kaigara printenv $KEY)
  if [ "$out" == "$VALUE" ]; then
    echo "ENV: { Round: ${ROUND}, Key: ${KEY}, Out: ${out} }"
  else
    echo "ERROR: { Round: ${ROUND}, Key: ${KEY}, Out: ${out} }"
    exit 1
  fi
done
