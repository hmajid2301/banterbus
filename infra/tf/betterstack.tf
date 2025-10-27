# # Website Monitor
# resource "betteruptime_monitor" "banterbus_website" {
#   provider = betterstack
#
#   url             = "https://banterbus.games"
#   monitor_type    = "status"
#   check_frequency = 60
#
#   email               = true
#   paused              = false
#   regions             = ["us", "eu"]
#   recovery_period     = 0
#   confirmation_period = 30
# }
#
# # API Health Monitor
# resource "betteruptime_monitor" "banterbus_api_health" {
#   provider = betterstack
#
#   url             = "https://banterbus.games/health"
#   monitor_type    = "status"
#   check_frequency = 30
#
#   email                 = true
#   paused                = false
#   regions               = ["us", "eu"]
#   recovery_period       = 0
#   confirmation_period   = 30
#   expected_status_codes = [200]
# }
#
# # Heartbeat Monitor for Application Health
# resource "betteruptime_heartbeat" "banterbus_app" {
#   provider = betterstack
#
#   name   = "banterbus-app-heartbeat"
#   period = 60
#   grace  = 30
#   email  = true
# }
#
# # Status Page
# resource "betteruptime_status_page" "banterbus" {
#   provider = betterstack
#
#   company_name = "Banter Bus"
#   company_url  = "https://banterbus.games"
#   contact_url  = "mailto:support@banterbus.games"
#
#   timezone     = "UTC"
#   subdomain    = "status-banterbus"
#   design       = "v2"
#   layout       = "vertical"
#   subscribable = true
# }
#
# # Status Page Sections
# resource "betteruptime_status_page_section" "website" {
#   provider = betterstack
#
#   status_page_id = betteruptime_status_page.banterbus.id
#   name           = "Website"
#   position       = 1
# }
#
# resource "betteruptime_status_page_section" "api" {
#   provider = betterstack
#
#   status_page_id = betteruptime_status_page.banterbus.id
#   name           = "API"
#   position       = 2
# }
#
# # Status Page Resources
# resource "betteruptime_status_page_resource" "website_monitor" {
#   provider = betterstack
#
#   status_page_id = betteruptime_status_page.banterbus.id
#   resource_id    = betteruptime_monitor.banterbus_website.id
#   resource_type  = "Monitor"
#   public_name    = "Website"
#
#   status_page_section_id = betteruptime_status_page_section.website.id
#   position               = 1
# }
#
# resource "betteruptime_status_page_resource" "api_health_monitor" {
#   provider = betterstack
#
#   status_page_id = betteruptime_status_page.banterbus.id
#   resource_id    = betteruptime_monitor.banterbus_api_health.id
#   resource_type  = "Monitor"
#   public_name    = "API Health"
#
#   status_page_section_id = betteruptime_status_page_section.api.id
#   position               = 1
# }
#
# # Outputs
# output "betterstack_heartbeat_url" {
#   description = "Heartbeat URL for Banterbus application"
#   value       = betteruptime_heartbeat.banterbus_app.url
# }
#
# output "betterstack_website_monitor_id" {
#   description = "Banterbus website monitor ID"
#   value       = betteruptime_monitor.banterbus_website.id
# }
#
# output "betterstack_api_health_monitor_id" {
#   description = "Banterbus API health monitor ID"
#   value       = betteruptime_monitor.banterbus_api_health.id
# }
#
# output "betterstack_status_page_url" {
#   description = "Banterbus status page URL"
#   value       = "https://${betteruptime_status_page.banterbus.subdomain}.betteruptime.com"
# }
#
# output "betterstack_status_page_subdomain" {
#   description = "Banterbus status page subdomain"
#   value       = betteruptime_status_page.banterbus.subdomain
# }

