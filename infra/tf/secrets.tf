# This file centralizes all secret management from OpenBao
# All secrets should be stored in OpenBao and read here

# GitLab secrets (already exists)
data "vault_kv_secret_v2" "gitlab_secrets" {
  mount = "kv"
  name  = "apps/gitlab"
}

# Cloudflare secrets (tunnel info already exists)
data "vault_kv_secret_v2" "cloudflare_secrets" {
  mount = "kv"
  name  = "infra/cloudflare"
}

# Cloudflare Terraform secrets (api_token and zone_id)
data "vault_kv_secret_v2" "cloudflare_tofu_secrets" {
  mount = "kv"
  name  = "infra/tofu"
}

# Database secrets (using existing postgres terraform path)
data "vault_kv_secret_v2" "database_secrets" {
  mount = "kv"
  name  = "infra/postgres/terraform"
}

# Grafana secrets
data "vault_kv_secret_v2" "grafana_secrets" {
  mount = "kv"
  name  = "infra/grafana"
}

# BetterStack secrets  
data "vault_kv_secret_v2" "betterstack_secrets" {
  mount = "kv"
  name  = "infra/betterstack"
}



# Local values to reference secrets cleanly
locals {
  # GitLab secrets
  gitlab_token                = data.vault_kv_secret_v2.gitlab_secrets.data["token"]
  gitlab_username             = data.vault_kv_secret_v2.gitlab_secrets.data["username"]
  gitlab_remote_state_address = var.remote_state_address # From terraform.tfvars since not in vault

  # Cloudflare secrets (all from OpenBao now)
  cloudflare_api_token = data.vault_kv_secret_v2.cloudflare_tofu_secrets.data["cloudflare_api_token"]
  cloudflare_zone_id   = data.vault_kv_secret_v2.cloudflare_tofu_secrets.data["cloudflare_zone_id"]
  tunnel_name          = data.vault_kv_secret_v2.cloudflare_secrets.data["tunnel_name"]
  tunnel_id            = data.vault_kv_secret_v2.cloudflare_secrets.data["tunnel_id"]
  tunnel_hostname      = "${local.tunnel_id}.cfargotunnel.com"

  # Database secrets (from existing postgres/terraform path)
  postgres_host     = data.vault_kv_secret_v2.database_secrets.data["host"]
  postgres_port     = tonumber(data.vault_kv_secret_v2.database_secrets.data["port"])
  postgres_username = data.vault_kv_secret_v2.database_secrets.data["username"]
  postgres_password = data.vault_kv_secret_v2.database_secrets.data["password"]
  postgres_database = "banterbus" # Default database name

  # Grafana secrets (from OpenBao)
  grafana_service_account_token = data.vault_kv_secret_v2.grafana_secrets.data["service_account_token"]

  # BetterStack secrets (from OpenBao)
  betterstack_api_token = data.vault_kv_secret_v2.betterstack_secrets.data["banterbus_api_token"]
}