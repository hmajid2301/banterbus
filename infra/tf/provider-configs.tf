# PostgreSQL Provider Configuration
provider "postgresql" {
  alias           = "homelab"
  host            = var.postgres_host
  port            = var.postgres_port
  database        = "postgres" # Connect to default database
  username        = var.postgres_username
  password        = var.postgres_password
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
  api_token = var.cloudflare_api_token
}

# Grafana Provider Configuration
provider "grafana" {
  alias                    = "homelab"
  url                      = "https://grafana.homelab.haseebmajid.dev"
  service_account_token    = var.grafana_service_account_token
}





