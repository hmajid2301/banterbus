# PostgreSQL Outputs
output "banterbus_dev_database" {
  description = "Banterbus dev database name"
  value       = postgresql_database.banterbus_dev.name
}

output "banterbus_prod_database" {
  description = "Banterbus prod database name"
  value       = postgresql_database.banterbus_prod.name
}

output "banterbus_dev_username" {
  description = "Banterbus dev database username"
  value       = postgresql_role.banterbus_dev.name
}

output "banterbus_prod_username" {
  description = "Banterbus prod database username"
  value       = postgresql_role.banterbus_prod.name
}

# OpenBao Secret Paths
output "banterbus_dev_secret_path" {
  description = "OpenBao secret path for dev environment"
  value       = vault_kv_secret_v2.banterbus_dev.name
}

output "banterbus_prod_secret_path" {
  description = "OpenBao secret path for prod environment"
  value       = vault_kv_secret_v2.banterbus_prod.name
}

# OpenBao Policy Names
output "banterbus_dev_policy" {
  description = "OpenBao policy name for dev access"
  value       = vault_policy.banterbus_dev.name
}

output "banterbus_prod_policy" {
  description = "OpenBao policy name for prod access"
  value       = vault_policy.banterbus_prod.name
}

# Cloudflare Tunnel Information
output "cloudflare_tunnel_name" {
  description = "Cloudflare tunnel name from OpenBao"
  value       = local.tunnel_name
  sensitive   = true
}

output "cloudflare_tunnel_id" {
  description = "Cloudflare tunnel ID from OpenBao"
  value       = local.tunnel_id
  sensitive   = true
}

output "cloudflare_tunnel_hostname" {
  description = "Generated Cloudflare tunnel hostname"
  value       = local.tunnel_hostname
  sensitive   = true
}