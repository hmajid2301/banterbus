# resource "sentry_organization" "banterbus" {
#   name = var.sentry_organization
#   slug = var.sentry_organization
# }

# resource "sentry_project" "banterbus_frontend" {
#   organization = sentry_organization.banterbus.slug
#   teams        = [sentry_team.banterbus_team.slug]
#   name         = "banterbus-frontend"
#   slug         = "banterbus-frontend"
#   platform     = "javascript"
# }

# resource "sentry_project" "banterbus_backend" {
#   organization = sentry_organization.banterbus.slug
#   teams        = [sentry_team.banterbus_team.slug]
#   name         = "banterbus-backend"
#   slug         = "banterbus-backend"
#   platform     = "go"
# }

# resource "sentry_team" "banterbus_team" {
#   organization = sentry_organization.banterbus.slug
#   name         = "banterbus-team"
#   slug         = "banterbus-team"
# }

# resource "sentry_key" "frontend_dsn" {
#   organization = sentry_organization.banterbus.slug
#   project      = sentry_project.banterbus_frontend.slug
#   name         = "frontend-key"
# }

# resource "sentry_key" "backend_dsn" {
#   organization = sentry_organization.banterbus.slug
#   project      = sentry_project.banterbus_backend.slug
#   name         = "backend-key"
# }

# output "sentry_frontend_dsn" {
#   value = sentry_key.frontend_dsn.dsn_public
# }

# output "sentry_backend_dsn" {
#   value = sentry_key.backend_dsn.dsn_secret
#   sensitive = true
# }