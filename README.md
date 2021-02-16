# Kaigara

Kaigara is a wrapper arround daemons. It helps to standardize the way daemons are started and configured in a cluster.

## Features

 * Fetch configuration from vault and inject in daemon environment
 * Support the storage of configuration files into vault
 * Push daemon STDOUT and STDERR to redis

## Quick start

```
export KAIGARA_REDIS_URL=redis://localhost:6379/0
export KAIGARA_VAULT_ADDR=http://127.0.0.1:8200
export KAIGARA_VAULT_TOKEN=s.ozytsgX1BcTQaR5Y07SAd2VE
export KAIGARA_APP_NAME=peatio
export KAIGARA_DEPLOYMENT_ID=opendax_uat
kagara service_command arguments...
```

    Note: You need to enable the kv and transit engine during the first time
    vault secrets enable kv -version=2
    vault secrets enable transit

**Warning**: You **must** enable Vault kv version 2 for Kaigara to function

## List existing secrets

To list existing app names, run:
```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID
```

To list existing scopes for an app name, run
```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME
```

## Read existing secrets

To read existing secrets for a given app name and scope, run:
```sh
vault read secret/data/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME/*scope* -format=yaml
```

## Bulk writing secrets to the SecretStore

To write secrets from the command line, save in a YAML file with a format similar to `secrets.yaml` and use `kaisave -f *filepath*`

**Warning**: All scopes to be used by a component **must** be initialized(e.g. `public: {}, private: {}, secret: {}`)

An example import file look similar to:
```yaml
secrets:
  global:
    scopes:
      public:
        key1: value1
        key2: [value2, value3]
        key3:
          key4: value4
          time:
            to: recover
      private:
        key1: value1
        key2: [value2, value3]
      secret:
        key1: value1
  peatio:
    scopes:
      public: {}
      private: {}
      secret:
        key1: value1
```
