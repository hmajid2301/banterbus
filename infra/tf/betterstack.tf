# resource "betterstack_monitor" "banterbus_website" {
#   url               = "https://banterbus.app"
#   monitor_type      = "status"
#   call             = false
#   sms              = false
#   email            = true
#   push             = false
#   team_wait        = 0
#   http_method      = "GET"
#   request_timeout  = 30
#   recovery_period  = 0
#   check_frequency  = 30
#   confirmations    = 1
#   follow_redirects = false
#   remember_cookies = false

#   request_headers = [
#     {
#       name  = "User-Agent"
#       value = "BetterStack"
#     }
#   ]

#   expected_status_codes = [200]
# }

# resource "betterstack_monitor" "banterbus_api_health" {
#   url               = "https://banterbus.app/health"
#   monitor_type      = "status"
#   call             = false
#   sms              = false
#   email            = true
#   push             = false
#   team_wait        = 0
#   http_method      = "GET"
#   request_timeout  = 30
#   recovery_period  = 0
#   check_frequency  = 30
#   confirmations    = 1
#   follow_redirects = false
#   remember_cookies = false

#   expected_status_codes = [200]
# }

# // BetterStack source for collecting logs
# resource "betterstack_source" "banterbus_logs" {
#   name        = "banterbus-logs"
#   platform    = "http"
#   retain_days = 7
# }

# // BetterStack source for collecting metrics
# resource "betterstack_source" "banterbus_metrics" {
#   name        = "banterbus-metrics"
#   platform    = "prometheus"
#   retain_days = 30
# }

# // BetterStack heartbeat for application health
# resource "betterstack_heartbeat" "banterbus_app" {
#   name        = "banterbus-app"
#   period      = 60
#   grace       = 300
#   call        = false
#   sms         = false
#   email       = true
#   push        = false
#   team_wait   = 0
#   sort        = 0
#   maintenance = []
# }

# output "betterstack_logs_token" {
#   value = betterstack_source.banterbus_logs.token
#   sensitive = true
# }

# output "betterstack_metrics_token" {
#   value = betterstack_source.banterbus_metrics.token
#   sensitive = true
# }

# output "betterstack_heartbeat_url" {
#   value = betterstack_heartbeat.banterbus_app.url
# }

# output "betterstack_website_monitor_id" {
#   value = betterstack_monitor.banterbus_website.id
# }

# output "betterstack_api_health_monitor_id" {
#   value = betterstack_monitor.banterbus_api_health.id
# }