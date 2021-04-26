# Kaigara

Kaigara is a wrapper arround daemons. It helps to standardize the way daemons are started and configured in a cluster.

## Features

 * Fetch configuration from Vault and inject in daemon environment
 * Support the storage of configuration files and env vars into Vault
 * Push daemon STDOUT and STDERR to redis

## Quick start

```sh
export KAIGARA_REDIS_URL=redis://localhost:6379/0
export KAIGARA_VAULT_ADDR=http://127.0.0.1:8200
export KAIGARA_VAULT_TOKEN=s.ozytsgX1BcTQaR5Y07SAd2VE
export KAIGARA_APP_NAME=peatio
export KAIGARA_DEPLOYMENT_ID=opendax_uat
# Optional - ignore global secret updates
export KAIGARA_IGNORE_GLOBAL=true

kagara service_command arguments...
```

    Note: You need to enable the kv and transit engine during the first time
    vault secrets enable kv -version=2
    vault secrets enable transit

**Warning**: You **must** enable Vault kv version 2 for Kaigara to function

## Manage secrets

To **list** existing **app names**, run:
```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID
```

To **list** existing scopes for an app name, run
```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME
```

To **read** existing secrets for a given app name and scope, run:
```sh
vault read secret/data/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME/*scope* -format=yaml
```

To **delete** existing secrets for a given app name and scope, run:
```sh
vault delete secret/data/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME/*scope*
```

### Bulk writing secrets to the SecretStore

To write secrets from the command line, save in a YAML file with a format similar to `secrets.yaml` and use `kaisave -f *filepath*`

**Warning**: All scopes to be used by a component **must** be initialized(e.g. `public: {}, private: {}, secret: {}`)

Make sure to wrap numeric and boolean values in quotes(e.g. `"4269"`, `"true"`), getting errors such as `interface{} is bool|json.Number|etc` is directly linked to unquoted values.

An example import file look similar to:
```yaml
secrets:
  global:
    scopes:
      private:
        global_key1: value1
        global_key2:
          - value2
          - value3
        global_key3:
          key4: value4
      public:
        global_key0: trustworthy
        global_key1: true
        global_key2:
          - value2
          - value3
        global_key3:
          key4: "1337"
          time:
            to: recover
        global_key1337: "1337"
      secret:
        global_key1: just a string
  peatio:
      scopes:
        private:
          key1: value1
          key2:
            - value2
            - value3
          key3:
            key4: value4
        public:
          key1: value1
          key2:
            - value2
            - value3
          key3:
            key4: value4
        secret:
          key1: value1
```

### Dump and output secrets from the SecretStore

To dump and output secrets from vault, use `kaidump -a <output.yaml>`


### Delete secret from the SecretStore

To delete secret from vault, use `kaidel -k <key name>`
