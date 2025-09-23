variable "remote_state_address" {
  type        = string
  description = "Gitlab remote state file address"
}

variable "username" {
  type        = string
  description = "Gitlab username to query remote state"
}

variable "access_token" {
  type        = string
  description = "GitLab access token to query remote state"
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token"
  type        = string
  sensitive   = true
}

variable "zone_id" {
  description = "Cloudflare Zone ID for banterbus.games domain"
  type        = string
}

# Cloudflare tunnel hostname is now dynamically generated from OpenBao
# See openbao.tf for the data source and local.tunnel_hostname

# variable "grafana_cloud_access_policy_token" {
#   description = "Grafana cloud access policy token"
#   type        = string
# }

variable "postgres_host" {
  description = "PostgreSQL host"
  type        = string
}

variable "postgres_port" {
  description = "PostgreSQL port"
  type        = number
  default     = 5432
}

variable "postgres_database" {
  description = "PostgreSQL database name"
  type        = string
  default     = "banterbus"
}

variable "postgres_username" {
  description = "PostgreSQL username"
  type        = string
}

variable "postgres_password" {
  description = "PostgreSQL password (deprecated - use cert auth)"
  type        = string
  sensitive   = true
  default     = ""
}


# variable "sentry_auth_token" {
#   description = "Sentry authentication token"
#   type        = string
#   sensitive   = true
# }

# variable "sentry_organization" {
#   description = "Sentry organization name"
#   type        = string
#   default     = "banterbus"
# }

# variable "betterstack_api_token" {
#   description = "API token for Better Stack"
#   type        = string
#   sensitive   = true
# }

# OpenBao Variables
variable "openbao_address" {
  description = "OpenBao server address"
  type        = string
  default     = "https://openbao.homelab.haseebmajid.dev"
}

variable "openbao_token" {
  description = "OpenBao authentication token"
  type        = string
  sensitive   = true
}

variable "grafana_service_account_token" {
  description = "Grafana service account token for dashboard management"
  type        = string
  sensitive   = true
}

variable "gitlab_project_id" {
  description = "GitLab project ID for banterbus"
  type        = string
  default     = "hmajid2301/banterbus"
}




# variable "alloy_endpoint" {
#   description = "Alloy homelab endpoint"
#   type        = string
#   default     = "alloy.homelab.haseebmajid.dev"
# }

