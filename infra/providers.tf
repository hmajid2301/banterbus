terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.5.0"
    }

    grafana = {
      source  = "grafana/grafana"
      version = ">= 2.9.0"
    }

    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.21"
    }

    sentry = {
      source  = "jianyuan/sentry"
      version = "~> 0.12.0"
    }

    betterstack = {
      source  = "BetterStackHQ/better-stack"
      version = "~> 0.5.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

provider "grafana" {
  alias                     = "cloud"
  cloud_access_policy_token = var.grafana_cloud_access_policy_token
}

provider "postgresql" {
  host            = var.postgres_host
  port            = var.postgres_port
  database        = var.postgres_database
  username        = var.postgres_username
  password        = var.postgres_password
  sslmode         = "require"
  connect_timeout = 15
}

provider "sentry" {
  token = var.sentry_auth_token
}

provider "betterstack" {
  api_token = var.betterstack_api_token
}