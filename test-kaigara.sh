#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root-token

COUNT=$1
if [ -z $COUNT ]; then
  COUNT=1
fi
FILE=./bin/kaigara
if [ ! -f "$FILE" ]; then
  CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/kaigara ./cmd/kaigara
fi
chmod +x ./assert-env.sh

docker-compose -f ./etc/backend.yml down -v
docker-compose -f ./etc/backend.yml up -d

vault secrets enable transit
go test -count=1 -run TestKaigaraPrintenv ./cmd/kaigara/

for i in $(seq 1 $COUNT); do
  sh ./assert-env.sh $i &
done
