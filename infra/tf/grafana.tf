# Apps folder for organizing application dashboards
resource "grafana_folder" "apps" {
  provider = grafana.homelab
  title    = "Apps"
  uid      = "apps-folder"
}

# Banter Bus dashboard in the Apps folder
resource "grafana_dashboard" "banterbus" {
  provider = grafana.homelab
  folder   = grafana_folder.apps.uid

  config_json = file("${path.module}/../../dashboards/grafana.json")
}



