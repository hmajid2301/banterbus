terraform {
  required_providers {
    betterstack = {
      source  = "BetterStackHQ/better-uptime"
      version = "~> 0.9.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "~> 17.0"
    }
    grafana = {
      source  = "grafana/grafana"
      version = "~> 3.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }

    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.21"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.4.0"
    }
  }
}

# BetterStack Provider Configuration
provider "betterstack" {
  api_token = local.betterstack_api_token
}

# Cloudflare Provider Configuration
provider "cloudflare" {
  api_token = local.cloudflare_api_token
}

# GitLab Provider Configuration
provider "gitlab" {
  token    = local.gitlab_token
  base_url = "https://gitlab.com/api/v4/"
}

# Grafana Provider Configuration
provider "grafana" {
  alias = "homelab"
  url   = "https://grafana.homelab.haseebmajid.dev"
  auth  = local.grafana_service_account_token
}

# Kubernetes Provider Configuration
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kubeconfig_context
}

# PostgreSQL Provider Configuration
provider "postgresql" {
  alias           = "homelab"
  host            = local.postgres_host
  port            = local.postgres_port
  database        = "postgres"
  username        = local.postgres_username
  password        = local.postgres_password
  sslmode         = "disable"
  connect_timeout = 15
  superuser       = true
}

# OpenBao/Vault Provider Configuration
provider "vault" {
  address          = var.openbao_address
  token            = var.openbao_token
  skip_child_token = true
}

# ============================================================================
# OUTPUTS
# ============================================================================

# Database Outputs
output "banterbus_dev_database" {
  description = "Banterbus dev database name"
  value       = postgresql_database.banterbus_dev.name
}

output "banterbus_prod_database" {
  description = "Banterbus prod database name"
  value       = postgresql_database.banterbus_prod.name
}

output "banterbus_dev_username" {
  description = "Banterbus dev database username"
  value       = postgresql_role.banterbus_dev.name
}

output "banterbus_prod_username" {
  description = "Banterbus prod database username"
  value       = postgresql_role.banterbus_prod.name
}

# OpenBao Outputs
output "banterbus_dev_secret_path" {
  description = "OpenBao secret path for dev environment"
  value       = vault_kv_secret_v2.banterbus_dev.name
}

output "banterbus_prod_secret_path" {
  description = "OpenBao secret path for prod environment"
  value       = vault_kv_secret_v2.banterbus_prod.name
}

output "banterbus_dev_policy" {
  description = "OpenBao policy name for dev access"
  value       = vault_policy.banterbus_dev.name
}

output "banterbus_prod_policy" {
  description = "OpenBao policy name for prod access"
  value       = vault_policy.banterbus_prod.name
}

# Cloudflare Outputs
output "cloudflare_tunnel_name" {
  description = "Cloudflare tunnel name from OpenBao"
  value       = local.tunnel_name
  sensitive   = true
}

output "cloudflare_tunnel_id" {
  description = "Cloudflare tunnel ID from OpenBao"
  value       = local.tunnel_id
  sensitive   = true
}

output "cloudflare_tunnel_hostname" {
  description = "Generated Cloudflare tunnel hostname"
  value       = local.tunnel_hostname
  sensitive   = true
}

# Grafana Outputs
output "banterbus_dashboard_url" {
  description = "URL to the Banter Bus dashboard in the Apps folder"
  value       = "https://grafana.homelab.haseebmajid.dev/d/${grafana_dashboard.banterbus.uid}/banter-bus"
}