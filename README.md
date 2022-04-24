# Kaigara

Kaigara is a wrapper arround daemons. It helps to standardize the way daemons are started and configured in a cluster.

## Features

 * Fetch configuration from secret storage and inject in daemon environment
 * Support the storage of configuration files and env vars into secret storage
 * Push daemon STDOUT and STDERR to redis
 * Restart process on outdated secrets

## Quick start

Kaigara supports two types of storage - Vault or SQL database, that can be used with `vault` and `sql` values respectively with env var below:

```bash
export KAIGARA_STORAGE_DRIVER=sql
```

If you choose Vault, then here are required **vault** vars for it to work:

```bash
export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
```

But if you choose SQL driver, then these **database** and **encryptor** vars should be set:

```bash
# As SQL drivers are supported postgres and mysql
export DATABASE_DRIVER=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASS=changeme
export KAIGARA_LOG_LEVEL=1

# As encryptors are supported transit (using Vault), aes and plaintext
export KAIGARA_ENCRYPTOR=transit

# If you use AES encryption method you need provide AES key
export KAIGARA_ENCRYPTOR_AES_KEY=changemechangeme

# If you use Vault transit encryption method you need to set Vault related vars
export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
```

After that in most situation you should set these **platform** vars as well:

```bash
# Your platform id used as secretspace in secret storage
export KAIGARA_DEPLOYMENT_ID=opendax_uat

# App names separated by comma
export KAIGARA_APP_NAME=peatio

# Scopes separated by comma
export KAIGARA_SCOPES=public,private,secret
```

If you are using `kaigara` cli, then you could also set:

```bash
# If you want to redirect logs to Redis channel
export KAIGARA_REDIS_URL=redis://localhost:6379/0

# If you want to ignore secrets in global app
export KAIGARA_IGNORE_GLOBAL=true
```

Example env vars are stored in [kaigara.env](./examples/kaigara.env).

## Manage secrets

### Using Vault

**Warning**: If you use Vault as a secret storage, then encryption using `transit` should be set:

```bash
vault secrets enable transit
```



To **list** existing **app names**, run:

```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID
```

To **list** existing **scopes** for an app name, run:
```sh
vault list secret/metadata/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME
```

To **read** existing secrets for a given app name and scope, run:
```sh
vault read secret/data/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME/$KAIGARA_SCOPES -format=yaml
```

To **delete** existing secrets for a given app name and scope, run:
```sh
vault delete secret/data/$KAIGARA_DEPLOYMENT_ID/$KAIGARA_APP_NAME/$KAIGARA_SCOPES
```

**Warning**: Commands above assume that vars `KAIGARA_APP_NAME` and `KAIGARA_SCOPES` are single (doesn't have commas).

### Using SQL

**Warning**: Queries below assume that you have active connection to Kaigara database, can run queries, and have enough permissions.

The name of Kaigara database is like `kaigara_$KAIGARA_DEPLOYMENT_ID`.

To **list** existing **app names**, run:

```sql
SELECT DISTINCT(app_name) FROM models;
```

To **list** existing **scopes** for an app name, run:

```sql
SELECT DISTINCT(scope) FROM models WHERE app_name = '*app_name*';
```

To **read** existing secrets for a given app name and scope, run:

```sql
SELECT value FROM models WHERE app_name = '*app_name*'AND scope = '*scope*';
```

To **delete** existing secrets for a given app name and scope, run:

```sql
DELETE FROM models WHERE app_name = '*app_name*'AND scope = '*scope*';
```

### Using kaicli

`kai` CLI tool encapsulates all next tools in one, so if you ran command `kaidump` before, then now you can run it in similar way `kai dump`.

Most of answers you can get just by running `kai -help` or `kai *cmd* -help`.

You can set `KAICONFIG` var in your shell to file path and store there **configuration of Kaigara** to reuse later.

For example, if create a file `~/.kaigara/kaiconf.yaml` with contents of [kaiconf.yaml](./examples/kaiconf.yaml), set var to its path and run:

```bash
kai dump
```

Then it will dump secrets from *peatio* app and scopes *public*, *private* and *secret*.

But if you run:

```bash
KAIGARA_SCOPES=private kai dump
```

Then will be dumped only *private* secrets from the same app.

With `kai` tool you can also redefine vars by passing values to parameters, so if we will continue with previous command:

```bash
KAIGARA_SCOPES=private kai dump -s public
```

Then will be dumped only *public* secrets from the same app.

### Bulk writing secrets to the SecretStore

To write secrets from the command line, save in a YAML file with a format similar to [secrets.yaml](./examples/secrets.yaml) and run:

```bash
kaisave -f *filepath*
```

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

To dump and output secrets from vault, run:

```bash
kaidump -o *outputs_path*
```

Make sure you've set `KAIGARA_SCOPES` env var before using `kaidump`.

### Delete secret from the SecretStore

To delete secret from vault, run:

```bash
kaidel -k *key_name*
```

### Print internal environment variables

To print all environment variables, run:

```bash
kaienv
```

To print exact environment variable, run:

```
kaienv *ENV_NAME*
```