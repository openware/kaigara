# Kaigara

Kaigara is an entrypoint/wrapper for commands, CLI's and beyond.
It enables teams to build components and deployments with improved configuration and observability our of the box.

## Features

 * Fetch configuration from secret storage and inject into target command environment
 * Support the storage of configuration files and env vars into secret storage(Vault KV, MySQL, PostgreSQL)
 * Publish target command STDOUT and STDERR to Redis
 * Restart subprocesses on configuration updates(allows for dynamic configs)
 * Create files on startup from env vars starting with `KNAME_`

See more in the [docs folder](./docs).

## Configuration

Kaigara supports two types of storage - Vault or SQL database, that can be used with `vault` and `sql` values respectively with env var below:

```sh
export KAIGARA_STORAGE_DRIVER=sql
```

If you choose Vault, here are the required vars:

```sh
export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
```

If you choose SQL driver, then these vars should be set:

```sh
# Supported SQL drivers are postgres and mysql
export DATABASE_DRIVER=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASS=changeme
export KAIGARA_LOG_LEVEL=1
```

Both storage drivers are created with **encryptor**, that is used to encrypt/decrypt vars in the secret scope:

```sh
# Supported encryptors are transit (using Vault Transit), aes and plaintext (default)
export KAIGARA_ENCRYPTOR=transit

# If you use AES encryption method, you need provide an AES key
export KAIGARA_ENCRYPTOR_AES_KEY=changemechangeme

# For Vault transit encryption method, use the following
export KAIGARA_VAULT_ADDR=http://localhost:8200
export KAIGARA_VAULT_TOKEN=changeme
```

After that in most situation you should set these **platform** vars as well:

```sh
# Your platform id used as secretspace in secret storage
export KAIGARA_DEPLOYMENT_ID=opendax_uat

# [OPTIONAL] App names separated by comma
export KAIGARA_APP_NAME=peatio

# Scopes separated by comma
export KAIGARA_SCOPES=public,private,secret
```

If you are using `kaigara` CLI, you could also set:

```sh
# If you want to redirect logs to a Redis channel
export KAIGARA_REDIS_URL=redis://localhost:6379/0

# If you want to ignore secrets in global app
export KAIGARA_IGNORE_GLOBAL=true
```

Example env vars are stored in [kaigara.env](./examples/kaigara.env).

## Manage secrets

### Vault

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

### SQL

**Warning**: Queries below assume that you have active connection to Kaigara database, can run queries, and have enough permissions.

The name of Kaigara database is like `kaigara_$KAIGARA_DEPLOYMENT_ID`.

To **list** existing **app names**, run:

```sql
SELECT DISTINCT(app_name) FROM data;
```

To **list** existing **scopes** for an app name, run:

```sql
SELECT DISTINCT(scope) FROM data WHERE app_name = '*app_name*';
```

To **read** existing secrets for a given app name and scope, run:

```sql
SELECT value FROM data WHERE app_name = '*app_name*'AND scope = '*scope*';
```

To **delete** existing secrets for a given app name and scope, run:

```sql
DELETE FROM data WHERE app_name = '*app_name*'AND scope = '*scope*';
```

### Encryptor

Encryptor is used only to encrypt/decrypt vars from `secret` scope.

If you use `plaintext` (default setting), then there is no encryption and you can read your secrets freely, but in the case of `transit` or `aes` encryption you won't be able to read their contents directly, you'd only see its encrypted version.

#### Transit

**Warning**: If you `transit` encryptor, make sure to enable Transit engine in Vault:

```bash
vault secrets enable transit
```

To **find out** whether Transit key exists or not:

```sh
vault list transit/keys | grep *deployment_id*_kaigara_*app_name*
```

To **create** a Transit key, run:

```sh
vault write -f transit/keys/*deployment_id*_kaigara_*app_name*
```

To **encrypt** a plain text string, run:

```sh
vault write transit/encrypt/*deployment_id*_kaigara_*app_name* -plaintext=*text*
```

To **decrypt** a cipher text string, run:
```sh
vault write transit/decrypt/*deployment_id*_kaigara_*app_name* -ciphertext=*text*
```

### AES

The AES encryptor type is implemented with GCM, that currently is not supported by `openssl` CLI tool.

If you need to debug or just encrypt/decrypt secrets in the same way as Kaigare does it, you can use something like [this](https://github.com/jforissier/aesgcm).

### Using kai CLI

`kai` CLI tool encapsulates all the previously separated tools(`kaidump`, `kaisave`, `kaidump`, `kaidel`) in one. For example, if you ran command `kaidump` before, now you can run it as `kai dump`.

If you're not sure about any subcommand's usage, run `kai -help` or `kai *cmd* -help`.

You can set `KAICONFIG` var in your shell to file path and store there **configuration of Kaigara** there to reuse later.

For example, if a file `~/.kaigara/kaiconf.yaml` with contents of [kaiconf.yaml](./examples/kaiconf.yaml) is created, set `KAICONFIG` to its path and run:

```bash
kai dump
```

It will dump secrets from *peatio* app and *public*, *private* and *secret* scopes, exactly as mentioned in the config.

But if you run:

```bash
KAIGARA_SCOPES=private kai dump
```

The env var would override the file config and only *private* secrets will be dumped from the configured app.

With `kai` tool you can also redefine vars by passing values to parameters, so if we will continue with previous command:

```bash
KAIGARA_SCOPES=private kai dump -s public
```

Then only *public* secrets will be dumped from the same app.

### Bulk writing secrets to the secret store

To write secrets from the command line, save in a YAML file with a format similar to [secrets.yaml](./examples/secrets.yaml) and run:

```bash
kai save -f *filepath*
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

### Dump and output configs

To dump and output secrets from the storage, run:

```bash
kai dump -o *outputs_path*
```

Make sure you've set `KAIGARA_SCOPES` env var before using `kaidump`.

### Delete configs

To delete configs from the storage, run:

```bash
kai del *app.scope.var*
```

For example, if you want to delete `finex_database_host` from `secret` scope in `finex` app, you should run:

```bash
kai del finex.secret.finex_database_host
```

You can also delete all entries from a scope:

```bash
kai del finex.secret.all
```

Or from the whole app:

```bash
kai del finex.all.all
```

Or even all present secrets from the current deployment ID:

```bash
kai del all.all.all
```

### Print internal environment variables

To print all environment variables including the ones loaded by Kaigara from the secret storage, run:

```bash
kai env
```

To print exact environment variable, run:

```
kai env *ENV_NAME*
```
