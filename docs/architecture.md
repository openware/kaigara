# Basic concepts

## Variables precedences

1. Defaults (Barong like)
2. Reading from a YAML file
3. Reading from environment variables
4. Reading passed cli parameters
5. Reading from remote secret storages and watching changes

## Design

Key-value secret storage with a policy per service to allow to fetch and edit

1. Kaigara inject configuration in environment
2. Kaigara monitor for new version of configuration
3. How do we edit configuration?

## Secret keys

Each platform has its own namespace specified by deployment id.

### SQL

If you use SQL driver, then Kaigara use separate database for each platform.

All data is stored within `data` table in this database. For example, in PostgreSQL this table would look like:

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

Each component also has its own namespace within platform's namespace with 3 scopes(`public`, `private`, `secret`) in Vault kv:

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

### Wrap secrets as files

`KFILE`-like secrets can be used for creating files based on their values.

For each file that you want to be created by Kaigara process you should create two secrets:

* `KFILE_*NAME*_PATH`  - path of file to create. If it has nested directories, Kaigara will ensure that all of them are created
* `KFILE_*NAME*_CONTENT` - base64 encoded content of file to create. This way you can create any files acceptable by string length.

Lets do **some practice**.

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

Yeah, you did it!