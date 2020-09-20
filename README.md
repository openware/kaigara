# Kaigara

Kaigara is a wrapper arround daemons. It helps to standardize the way daemons are started and configured in a cluster.

## Features

 * Fetch congfiguration from vault and inject in daemon environment
 * Push daemon STDOUT and STDERR to redis

## Quick start

```
export KAIGARA_REDIS_URL=redis://localhost:6379/0
export KAIGARA_VAULT_ADDR=http://127.0.0.1:8200
export KAIGARA_VAULT_TOKEN=s.ozytsgX1BcTQaR5Y07SAd2VE
export KAIGARA_VAULT_CONFIG_PATH=cluster/your_service_name
export KAIGARA_SERVICE_NAME=YourServiceName
kagara service_command arguments...
```

## TODO

 * Detects configuration changes and apply by restarting the daemon with new environment
 * Handle a command message from redis to restart the daemon
