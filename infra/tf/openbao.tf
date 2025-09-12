# Store Banterbus dev secrets in OpenBao
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

# Store Banterbus prod secrets in OpenBao
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

# Create OpenBao policy for Banterbus dev access (for Kubernetes)
resource "vault_policy" "banterbus_dev" {
  name = "banterbus-dev"

  policy = <<EOT
# Allow reading banterbus dev secrets
path "kv/data/apps/dev/banterbus" {
  capabilities = ["read"]
}

path "kv/metadata/apps/dev/banterbus" {
  capabilities = ["read"]
}
EOT
}

# Create OpenBao policy for Banterbus prod access (for Kubernetes)
resource "vault_policy" "banterbus_prod" {
  name = "banterbus-prod"

  policy = <<EOT
# Allow reading banterbus prod secrets
path "kv/data/apps/prod/banterbus" {
  capabilities = ["read"]
}

path "kv/metadata/apps/prod/banterbus" {
  capabilities = ["read"]
}
EOT
}

# Read Cloudflare tunnel information from OpenBao
data "vault_kv_secret_v2" "cloudflare_tunnel" {
  mount = "kv"
  name  = "infra/cloudflare"
}

# Extract tunnel information
locals {
  tunnel_name = data.vault_kv_secret_v2.cloudflare_tunnel.data["tunnel_name"]
  tunnel_id   = data.vault_kv_secret_v2.cloudflare_tunnel.data["tunnel_id"]
  tunnel_hostname = "tunnel-${local.tunnel_name}-${substr(local.tunnel_id, 0, 8)}.cfargotunnel.com"
}

