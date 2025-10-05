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

variable "gitlab_project_id" {
  description = "GitLab project ID for banterbus"
  type        = string
  default     = "hmajid2301/banterbus"
}

variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
  default     = "dev"
}

# Kubernetes configuration variables
variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = null
}

variable "status_page_domain" {
  description = "Custom domain for status page (optional)"
  type        = string
  default     = ""
}

variable "admin_phone_number" {
  description = "Admin phone number for SMS alerts"
  type        = string
  default     = ""
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token"
  type        = string
  default     = null
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID"
  type        = string
  default     = null
}

variable "grafana_service_account_token" {
  description = "Grafana service account token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "betterstack_api_token" {
  description = "BetterStack API token"
  type        = string
  default     = ""
  sensitive   = true
}


