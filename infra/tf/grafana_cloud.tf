resource "grafana_cloud_stack" "banterbus_stack" {
  provider = grafana.cloud

  name        = "BanterBus Grafana Cloud"
  slug        = "banterbus"
  region_slug = "eu"
  description = "Production monitoring stack for BanterBus"
}

resource "grafana_cloud_stack_service_account" "terraform_sa" {
  provider   = grafana.cloud
  stack_slug = grafana_cloud_stack.banterbus_stack.slug

  name        = "terraform-manager"
  role        = "Admin"
  is_disabled = false
}

resource "grafana_cloud_stack_service_account" "alloy_sa" {
  provider   = grafana.cloud
  stack_slug = grafana_cloud_stack.banterbus_stack.slug

  name        = "alloy-homelab"
  role        = "MetricsPublisher"
  is_disabled = false
}

resource "grafana_cloud_access_policy" "alloy_policy" {
  provider = grafana.cloud

  region = grafana_cloud_stack.banterbus_stack.region_slug
  name   = "alloy-homelab-policy"
  scopes = ["metrics:write", "logs:write", "traces:write"]

  realm {
    type       = "org"
    identifier = grafana_cloud_stack.banterbus_stack.org_id
  }
}

output "grafana_cloud_prometheus_url" {
  value = grafana_cloud_stack.banterbus_stack.prometheus_remote_write_endpoint
}

output "grafana_cloud_loki_url" {
  value = grafana_cloud_stack.banterbus_stack.logs_url
}

output "grafana_cloud_tempo_url" {
  value = grafana_cloud_stack.banterbus_stack.otlp_url
}

output "grafana_cloud_stack_id" {
  value = grafana_cloud_stack.banterbus_stack.id
}

output "alloy_service_account_token" {
  value = grafana_cloud_stack_service_account.alloy_sa.key
  sensitive = true
}