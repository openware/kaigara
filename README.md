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
    vault secrets enable kv
    vault secrets enable transit

## Store an environment variable into Vault

The variable name will be upper cased.
```
vault kv put secret/cluster/your_service_name your_var=your_value
```

## Store a configuration file into Vault

```
vault kv patch secret/cluster/your_service_name kfile_config_path=config.json
vault kv patch secret/cluster/your_service_name kfile_config_content='{"app":"example"}'
```

## TODO

 * Detects configuration changes and apply by restarting the daemon with new environment
 * Handle a command message from redis to restart the daemon
