# Configuration

## Variables precedences

1. Defaults (Barong like)
2. Reading from a YAML file
3. Reading from environment variables
4. Reading passed cli parameters
5. Reading from remote config systems vault and watching changes

## Design

Vault kv v2 storage with a policy per service to allow to fetch and edit

1. Kaigara inject configuration in environment
2. Kaigara monitor for new version of configuration
3. How do we edit configuration?

## Secrets keys

Each component has its own namespace/key with 3 scopes(`public`, `private`, `secret`) in Vault kv:
  - kv/#{platform_id}/peatio/{public,private,secret}
  - kv/#{platform_id}/barong/{public,private,secret}
  - kv/#{platform_id}/finex/{public,private,secret}
  - kv/#{platform_id}/hd-wallet/{public,private,secret}

e.g. `kv/yellow/peatio/public` has:
```yaml
data:
  min_deposit_level: 1
  ...
```

Public data is exposed to the frontend via `env.js`.
Private data is passed to component environment vars and is available to view and edit from Tower.
Secret data is passed to component environment vars and is available `only to edit` from Tower.

## Edit configuration

Sonic endpoints:

1. GET public env.js returns an aggregated entries from public keys
2. GET admin api to return to admin public, private and **masked** secrets
3. POST admin api to push new configuration entries

:warning: Vault API doesn't allow to update one entries in a kv

Sonic has ability to read all data scopes(`public`, `private`, `secret`) but it can **only encrypt** `secret` data

Each component has its own Vault transit key for data encryption/decryption, the naming is as follows:
```
transit/{platform_id}_kaigara_{component}
```

## Watch changes

- Use vault kv store v2 or store the updated_at in ms
- Poll the last version
- If different restart the application