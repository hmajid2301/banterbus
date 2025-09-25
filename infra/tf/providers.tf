terraform {
  required_providers {
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "~> 1.21"
    }

    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.4.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.6.0"
    }

    sentry = {
      source  = "jianyuan/sentry"
      version = "~> 0.12.0"
    }

    betterstack = {
      source  = "BetterStackHQ/better-uptime"
      version = "~> 0.9.0"
    }

    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }

    grafana = {
      source  = "grafana/grafana"
      version = "~> 3.0"
    }

    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "~> 17.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }

    
  }
}

