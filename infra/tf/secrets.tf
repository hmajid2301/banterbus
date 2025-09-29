# This file centralizes all secret management from OpenBao
# All secrets should be stored in OpenBao and read here

# Infrastructure secrets (tokens, API keys, etc.)
data "vault_kv_secret_v2" "infra_secrets" {
  mount = "kv"
  name  = "infra"
}

# GitLab secrets (already exists but centralizing here)
data "vault_kv_secret_v2" "gitlab_secrets" {
  mount = "kv"
  name  = "apps/gitlab"
}

# Cloudflare secrets (tunnel info already exists, adding API token)
data "vault_kv_secret_v2" "cloudflare_secrets" {
  mount = "kv"
  name  = "infra/cloudflare"
}

# Database secrets
data "vault_kv_secret_v2" "database_secrets" {
  mount = "kv"
  name  = "infra/database"
}

# Grafana secrets
data "vault_kv_secret_v2" "grafana_secrets" {
  mount = "kv"
  name  = "infra/grafana"
}

# Local values to reference secrets cleanly
locals {
  # GitLab secrets
  gitlab_token                = data.vault_kv_secret_v2.gitlab_secrets.data["token"]
  gitlab_username             = data.vault_kv_secret_v2.gitlab_secrets.data["username"]
  gitlab_remote_state_address = data.vault_kv_secret_v2.gitlab_secrets.data["remote_state_address"]

  # Cloudflare secrets
  cloudflare_api_token = data.vault_kv_secret_v2.cloudflare_secrets.data["api_token"]
  cloudflare_zone_id   = data.vault_kv_secret_v2.cloudflare_secrets.data["zone_id"]
  tunnel_name          = data.vault_kv_secret_v2.cloudflare_secrets.data["tunnel_name"]
  tunnel_id            = data.vault_kv_secret_v2.cloudflare_secrets.data["tunnel_id"]
  tunnel_hostname      = "${local.tunnel_id}.cfargotunnel.com"

  # Database secrets
  postgres_host     = data.vault_kv_secret_v2.database_secrets.data["host"]
  postgres_port     = tonumber(data.vault_kv_secret_v2.database_secrets.data["port"])
  postgres_username = data.vault_kv_secret_v2.database_secrets.data["username"]
  postgres_password = data.vault_kv_secret_v2.database_secrets.data["password"]
  postgres_database = data.vault_kv_secret_v2.database_secrets.data["database"]

  # Grafana secrets
  grafana_service_account_token = data.vault_kv_secret_v2.grafana_secrets.data["service_account_token"]
}