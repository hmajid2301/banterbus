resource "cloudflare_dns_record" "root" {
  zone_id = var.zone_id
  name    = "@"
  type    = "CNAME"
  content = "banterbus.fly.dev"
  ttl     = 1
  proxied = false
}

resource "cloudflare_dns_record" "wildcard" {
  zone_id = var.zone_id
  name    = "*"
  type    = "CNAME"
  content = "banterbus.fly.dev"
  ttl     = 1
  proxied = false
}

# Dev APEX record (dev.banterbus.app)
resource "cloudflare_dns_record" "dev_apex" {
  zone_id = var.zone_id
  name    = "dev"
  type    = "CNAME"
  content = "banterbus-dev.fly.dev"
  ttl     = 1
  proxied = false
}

# Dev WILDCARD (*.dev.banterbus.app)
resource "cloudflare_dns_record" "dev_wildcard" {
  zone_id = var.zone_id
  name    = "*.dev"
  type    = "CNAME"
  content = "banterbus-dev.fly.dev"
  ttl     = 1
  proxied = false
}

# ACME challenge records for SSL certificates
resource "cloudflare_dns_record" "acme_challenge" {
  zone_id = var.zone_id
  name    = "_acme-challenge"
  type    = "CNAME"
  content = "banterbus.app.fly.dev"
  ttl     = 1
  proxied = false
}

resource "cloudflare_dns_record" "acme_challenge_dev" {
  zone_id = var.zone_id
  name    = "_acme-challenge.dev"
  type    = "CNAME"
  content = "dev.banterbus.app.fly.dev"
  ttl     = 1
  proxied = false
}