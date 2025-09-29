# Configure OpenBao Kubernetes Auth Backend
# This configures the backend to validate service account tokens

# Get the service account JWT token
resource "kubernetes_secret_v1" "vault_auth_token" {
  metadata {
    name      = "vault-auth-token"
    namespace = "default"
    annotations = {
      "kubernetes.io/service-account.name" = "vault-auth"
    }
  }
  type = "kubernetes.io/service-account-token"
}

# Configure the Kubernetes auth method backend
resource "vault_kubernetes_auth_backend_config" "k8s_config" {
  backend         = data.vault_auth_backend.kubernetes.path
  kubernetes_host = "https://kubernetes.default.svc.cluster.local:443"

  # Use the service account token for token review
  token_reviewer_jwt = kubernetes_secret_v1.vault_auth_token.data["token"]

  # Use the cluster CA certificate
  kubernetes_ca_cert = kubernetes_secret_v1.vault_auth_token.data["ca.crt"]

  # Disable local CA verification since we're providing the cert
  disable_local_ca_jwt = false
}