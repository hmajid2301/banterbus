# Use existing Kubernetes auth method (already configured)
# No need to create auth backend - it already exists

# Data source to reference existing kubernetes auth backend
data "vault_auth_backend" "kubernetes" {
  path = "kubernetes"
}

# Update existing role to include banterbus policies
resource "vault_kubernetes_auth_backend_role" "k8s_auth_role_update" {
  backend                          = data.vault_auth_backend.kubernetes.path
  role_name                        = "k8s-auth-role"
  bound_service_account_names      = ["openbao-auth", "banterbus"]
  bound_service_account_namespaces = ["apps", "default", "dev", "flux-system", "prod", "tailscale"]
  token_policies                   = [vault_policy.banterbus_dev.name, vault_policy.banterbus_prod.name, "default"]
  token_ttl                        = 3600
  token_max_ttl                    = 86400
}

