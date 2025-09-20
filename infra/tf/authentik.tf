# Remove the problematic data sources and locals for now
# We'll reference the otel-collector app directly by its known UUID from the plan output

resource "authentik_provider_oauth2" "banterbus_otel_dev" {
  name          = "banterbus-otel-dev-provider"
  client_id     = "banterbus-otel-dev"
  client_secret = random_password.banterbus_otel_dev_secret.result
  client_type   = "confidential"

  authorization_flow = data.authentik_flow.default_authorization_flow.id
  invalidation_flow  = data.authentik_flow.default_invalidation_flow.id
}

resource "authentik_provider_oauth2" "banterbus_otel_prod" {
  name          = "banterbus-otel-prod-provider"
  client_id     = "banterbus-otel-prod"
  client_secret = random_password.banterbus_otel_prod_secret.result
  client_type   = "confidential"

  authorization_flow = data.authentik_flow.default_authorization_flow.id
  invalidation_flow  = data.authentik_flow.default_invalidation_flow.id
}

resource "authentik_provider_oauth2" "banterbus_otel_ci" {
  name          = "banterbus-otel-ci-provider"
  client_id     = "banterbus-otel-ci"
  client_secret = random_password.banterbus_otel_ci_secret.result
  client_type   = "confidential"

  authorization_flow = data.authentik_flow.default_authorization_flow.id
  invalidation_flow  = data.authentik_flow.default_invalidation_flow.id
}

resource "random_password" "banterbus_otel_dev_secret" {
  length  = 64
  special = true
}

resource "random_password" "banterbus_otel_prod_secret" {
  length  = 64
  special = true
}

resource "random_password" "banterbus_otel_ci_secret" {
  length  = 64
  special = true
}

resource "authentik_user" "banterbus_otel_dev" {
  username = "banterbus-otel-dev"
  name     = "BanterBus OTEL Client (Dev)"
  email    = "banterbus-otel-dev@haseebmajid.dev"
  is_active = true
  type     = "service_account"
}

resource "authentik_user" "banterbus_otel_prod" {
  username = "banterbus-otel-prod"
  name     = "BanterBus OTEL Client (Prod)"
  email    = "banterbus-otel-prod@haseebmajid.dev"
  is_active = true
  type     = "service_account"
}

resource "authentik_user" "banterbus_otel_ci" {
  username = "banterbus-otel-ci"
  name     = "BanterBus OTEL Client (GitLab CI)"
  email    = "banterbus-otel-ci@haseebmajid.dev"
  is_active = true
  type     = "service_account"
}

resource "authentik_application" "banterbus_otel_dev" {
  name  = "banterbus-otel-dev"
  slug  = "banterbus-otel-dev"
  group = "BanterBus"

  protocol_provider = authentik_provider_oauth2.banterbus_otel_dev.id

  meta_description = "BanterBus OTEL Client Authentication (Dev)"
  meta_publisher   = "BanterBus"
}

resource "authentik_application" "banterbus_otel_prod" {
  name  = "banterbus-otel-prod"
  slug  = "banterbus-otel-prod"
  group = "BanterBus"

  protocol_provider = authentik_provider_oauth2.banterbus_otel_prod.id

  meta_description = "BanterBus OTEL Client Authentication (Prod)"
  meta_publisher   = "BanterBus"
}

resource "authentik_application" "banterbus_otel_ci" {
  name  = "banterbus-otel-ci"
  slug  = "banterbus-otel-ci"
  group = "BanterBus"

  protocol_provider = authentik_provider_oauth2.banterbus_otel_ci.id

  meta_description = "BanterBus OTEL Client Authentication (GitLab CI)"
  meta_publisher   = "BanterBus"
}

resource "authentik_group" "otel_access" {
  name = "otel-access"
  users = [
    authentik_user.banterbus_otel_dev.id,
    authentik_user.banterbus_otel_prod.id,
    authentik_user.banterbus_otel_ci.id,
  ]
}

resource "authentik_policy_binding" "otel_collector_access" {
  target = "65345fa1-c65f-4d6b-b119-a50151786fdf"  # otel-collector app UUID from plan output
  group  = authentik_group.otel_access.id
  order  = 0
}

data "authentik_flow" "default_authorization_flow" {
  slug = "default-authentication-flow"
}

data "authentik_flow" "default_invalidation_flow" {
  slug = "default-invalidation-flow"
}
