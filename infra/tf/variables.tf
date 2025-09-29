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

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = null
}

variable "kubeconfig_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = null
}
