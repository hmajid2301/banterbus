# PostgreSQL Provider Configuration
provider "postgresql" {
  alias           = "homelab"
  host            = local.postgres_host
  port            = local.postgres_port
  database        = "postgres" # Connect to default database
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

# Cloudflare Provider Configuration
provider "cloudflare" {
  api_token = local.cloudflare_api_token
}

# Grafana Provider Configuration
provider "grafana" {
  alias                 = "homelab"
  url                   = "https://grafana.homelab.haseebmajid.dev"
  service_account_token = local.grafana_service_account_token
}

# GitLab Provider Configuration
provider "gitlab" {
  token    = local.gitlab_token
  base_url = "https://gitlab.com/api/v4/"
}

# Kubernetes Provider Configuration
provider "kubernetes" {
  config_path    = var.kubeconfig_path != null ? var.kubeconfig_path : "~/.kube/config"
  config_context = var.kubeconfig_context != null ? var.kubeconfig_context : null
}





