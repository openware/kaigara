# View the kv v2 data
path "secret/data/deployment_id/*" {
  capabilities = ["read", "list"]
}

# View the kv v2 metadata
path "secret/metadata/deployment_id/*" {
  capabilities = ["read", "list"]
}

# Manage the transit secrets engine
path "transit/keys/deployment_id_kaigara_*" {
  capabilities = ["create", "update", "read", "list"]
}

# Encrypt secrets data
path "transit/encrypt/deployment_id_kaigara_*" {
  capabilities = ["create", "read", "update"]
}

# Decrypt secrets data
path "transit/decrypt/deployment_id_kaigara_*" {
  capabilities = ["create", "read", "update"]
}

# Renew tokens
path "auth/token/renew" {
  capabilities = ["update"]
}

# Lookup tokens
path "auth/token/lookup" {
  capabilities = ["update"]
}
