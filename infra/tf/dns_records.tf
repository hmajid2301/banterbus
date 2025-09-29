# Cloudflare Tunnel DNS Records for banterbus.games
# Using single tunnel for both dev and prod environments
# Tunnel hostname is dynamically generated from OpenBao data

# Production domain (banterbus.games)
resource "cloudflare_record" "prod_apex" {
  zone_id = local.cloudflare_zone_id
  name    = "@"
  type    = "CNAME"
  content = local.tunnel_hostname # Dynamically generated from OpenBao
  ttl     = 1
  proxied = true # Enable Cloudflare proxy for SSL and performance
}

# Production wildcard (*.banterbus.games)
resource "cloudflare_record" "prod_wildcard" {
  zone_id = local.cloudflare_zone_id
  name    = "*"
  type    = "CNAME"
  content = local.tunnel_hostname
  ttl     = 1
  proxied = true
}

# Development domain (dev.banterbus.games)
resource "cloudflare_record" "dev_apex" {
  zone_id = local.cloudflare_zone_id
  name    = "dev"
  type    = "CNAME"
  content = local.tunnel_hostname # Same tunnel for dev
  ttl     = 1
  proxied = true
}

# Development wildcard (*.dev.banterbus.games)
resource "cloudflare_record" "dev_wildcard" {
  zone_id = local.cloudflare_zone_id
  name    = "*.dev"
  type    = "A"
  content = "5.75.159.214"
  ttl     = 1
  proxied = true
}

# ACME challenge records for SSL certificates (if using cert-manager)
resource "cloudflare_record" "acme_challenge" {
  zone_id = local.cloudflare_zone_id
  name    = "_acme-challenge"
  type    = "CNAME"
  content = local.tunnel_hostname
  ttl     = 1
  proxied = false # ACME challenges should not be proxied
}

resource "cloudflare_record" "acme_challenge_dev" {
  zone_id = local.cloudflare_zone_id
  name    = "_acme-challenge.dev"
  type    = "CNAME"
  content = local.tunnel_hostname
  ttl     = 1
  proxied = false
}

