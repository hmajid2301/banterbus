resource "vault_kv_secret_v2" "banterbus_dev" {
  mount = "kv"
  name  = "apps/dev/banterbus"

  data_json = jsonencode({
    db_password = random_password.banterbus_dev.result
    db_username = postgresql_role.banterbus_dev.name
  })

  depends_on = [
    random_password.banterbus_dev,
    postgresql_role.banterbus_dev
  ]
}

resource "vault_kv_secret_v2" "banterbus_prod" {
  mount = "kv"
  name  = "apps/prod/banterbus"

  data_json = jsonencode({
    db_password = random_password.banterbus_prod.result
    db_username = postgresql_role.banterbus_prod.name
  })

  depends_on = [
    random_password.banterbus_prod,
    postgresql_role.banterbus_prod
  ]
}

resource "vault_policy" "banterbus_dev" {
  name = "banterbus-dev"

  policy = <<EOT
path "kv/data/apps/dev/banterbus" {
  capabilities = ["read"]
}

path "kv/metadata/apps/dev/banterbus" {
  capabilities = ["read"]
}
EOT
}

resource "vault_policy" "banterbus_prod" {
  name = "banterbus-prod"

  policy = <<EOT
path "kv/data/apps/prod/banterbus" {
  capabilities = ["read"]
}

path "kv/metadata/apps/prod/banterbus" {
  capabilities = ["read"]
}
EOT
}



# Cloudflare tunnel data and locals moved to secrets.tf

