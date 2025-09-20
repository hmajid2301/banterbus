resource "vault_kv_secret_v2" "banterbus_dev" {
  mount = "kv"
  name  = "apps/dev/banterbus"

  data_json = jsonencode({
    db_password = random_password.banterbus_dev.result
    db_username = postgresql_role.banterbus_dev.name
    otel_client_id = authentik_provider_oauth2.banterbus_otel_dev.client_id
    otel_client_secret = authentik_provider_oauth2.banterbus_otel_dev.client_secret
    otel_token_url = "https://authentik.haseebmajid.dev/application/o/token/"
    otel_issuer = "https://authentik.haseebmajid.dev/application/o/otel-collector/"
    otel_scopes = "openid"
  })

  depends_on = [
    random_password.banterbus_dev,
    postgresql_role.banterbus_dev,
    authentik_provider_oauth2.banterbus_otel_dev
  ]
}

resource "vault_kv_secret_v2" "banterbus_prod" {
  mount = "kv"
  name  = "apps/prod/banterbus"

  data_json = jsonencode({
    db_password = random_password.banterbus_prod.result
    db_username = postgresql_role.banterbus_prod.name
    otel_client_id = authentik_provider_oauth2.banterbus_otel_prod.client_id
    otel_client_secret = authentik_provider_oauth2.banterbus_otel_prod.client_secret
    otel_token_url = "https://authentik.haseebmajid.dev/application/o/token/"
    otel_issuer = "https://authentik.haseebmajid.dev/application/o/otel-collector/"
    otel_scopes = "openid"
  })

  depends_on = [
    random_password.banterbus_prod,
    postgresql_role.banterbus_prod,
    authentik_provider_oauth2.banterbus_otel_prod
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



data "vault_kv_secret_v2" "cloudflare_tunnel" {
  mount = "kv"
  name  = "infra/cloudflare"
}

locals {
  tunnel_name = data.vault_kv_secret_v2.cloudflare_tunnel.data["tunnel_name"]
  tunnel_id   = data.vault_kv_secret_v2.cloudflare_tunnel.data["tunnel_id"]
  tunnel_hostname = "${local.tunnel_id}.cfargotunnel.com"
}

