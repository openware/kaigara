# Basic concepts

## Variables precedences

1. Defaults (Barong like)
2. Reading from a YAML file
3. Reading from environment variables
4. Reading passed cli parameters
5. Reading from remote secret storages and watching changes

## Design

Kaigara provides a key-value config/secret storage with access to each app name limited to its user/token

1. Kaigara inject configuration in environment
2. Kaigara monitor for new version of configuration
3. How do we edit configuration?

## Secret keys

Each platform has its own namespace specified by deployment id.

### SQL

If you use SQL driver, Kaigara would use a separate database per deployment ID.

All data is stored in `data` table inside this database. For example, in PostgreSQL this table would look like:

```sql
CREATE TABLE data (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    app_name TEXT,
    scope TEXT,
    value JSON,
    version BIGINT
)
```

### Vault	

Each component also has its own namespace defined by its `app_name` with 3 scopes(`public`, `private`, `secret`) in [Vault KV](https://www.vaultproject.io/docs/secrets/kv):

  - kv/#{platform_id}/peatio/{public,private,secret}
  - kv/#{platform_id}/barong/{public,private,secret}
  - kv/#{platform_id}/finex/{public,private,secret}
  - kv/#{platform_id}/hd-wallet/{public,private,secret}

For example, `kv/yellow/peatio/public` has:
```yaml
min_deposit_level: 1
default_theme: dark
```

## Features

### Write secrets to files

Config vars starting with `KFILE` would be written to files upon Kaigara startup.

For each file that you want to be created by Kaigara process you should create two secrets:

* `KFILE_*NAME*_PATH`  - path of the file to be created. If it contains nested directories, Kaigara will ensure that all of them are created
* `KFILE_*NAME*_CONTENT` - base64 encoded content of the file to create. Any content would work as long as it can be put into an env var.

Let's do **some practice**.

First of all, create a file called `temp.txt`:

```bash
echo "you did it" > temp.txt
```

After that encode its contents to base64 format:

```
cat temp.txt | base64 -w0
```

This will output a string `eW91IGRpZCBpdAo=`, which is now in right format to insert in secret.

Next create a file called `secrets.yaml` with set `KFILE` secrets:

```bash
echo '
secrets:
  some_app:
    scopes:
      public:
        kfile_temp_path: new_temp.txt
        kfile_temp_content: eW91IGRpZCBpdAo=
' > secrets.yaml
```

And save it in the secret storage (this assumes, that you've already set configuration):

```bash
kai save -f secrets.yaml
```

Now you can run `kaigara` with no-daemon command (you don't want wait, do you?:):

```bash
KAIGARA_APP_NAME=some_app kaigara echo "just run"
```

And finally view the contents of newly created file:

```bash
cat new_file.txt
```

![GREAT SUCCESS](https://media.giphy.com/media/a0h7sAqON67nO/giphy.gif)

Yeah, you did it!
