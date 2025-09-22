terraform {
  required_providers {
    grafana = {
      source  = "grafana/grafana"
      version = "~> 3.0"
    }
  }
}

# Configure the Grafana provider for your homelab instance
provider "grafana" {
  alias               = "homelab"
  url                 = "https://grafana.homelab.haseebmajid.dev"
  service_account_token = var.grafana_service_account_token
}

# Create the apps/banterbus folder
resource "grafana_folder" "banterbus" {
  provider = grafana.homelab
  title    = "banterbus"
  uid      = "banterbus-folder"
}

# Deploy the BanterBus dashboard
resource "grafana_dashboard" "banterbus" {
  provider = grafana.homelab
  folder   = grafana_folder.banterbus.uid
  
  config_json = file("${path.module}/../../dashboards/grafana.json")
}

# Output the dashboard URL
output "banterbus_dashboard_url" {
  value = "https://grafana.homelab.haseebmajid.dev/d/${grafana_dashboard.banterbus.uid}/banterbus"
}