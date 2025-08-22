locals {
  alloy_config = templatefile("${path.module}/templates/alloy.river.tpl", {
    prometheus_url = grafana_cloud_stack.banterbus_stack.prometheus_remote_write_endpoint
    loki_url       = grafana_cloud_stack.banterbus_stack.logs_url
    tempo_url      = grafana_cloud_stack.banterbus_stack.otlp_url
    stack_id       = grafana_cloud_stack.banterbus_stack.id
    api_key        = grafana_cloud_stack_service_account.alloy_sa.key
  })
}

output "alloy_config" {
  value = local.alloy_config
  description = "Alloy configuration for connecting to Grafana Cloud"
}