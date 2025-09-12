# Generate secure random passwords for dev/prod environments
resource "random_password" "banterbus_dev" {
  length  = 32
  special = true
}

resource "random_password" "banterbus_prod" {
  length  = 32
  special = true
}

# Create banterbus dev user with modern auth settings
resource "postgresql_role" "banterbus_dev" {
  provider = postgresql.homelab
  name     = "banterbus_dev"
  login    = true
  password = random_password.banterbus_dev.result
  # Modern PostgreSQL security settings
  connection_limit    = 50
  valid_until        = "infinity"
  create_database    = false
  create_role        = false
  inherit            = true
  replication        = false
  superuser          = false
}

# Create banterbus prod user with modern auth settings
resource "postgresql_role" "banterbus_prod" {
  provider = postgresql.homelab
  name     = "banterbus_prod"
  login    = true
  password = random_password.banterbus_prod.result
  connection_limit    = 100
  valid_until        = "infinity"
  create_database    = false
  create_role        = false
  inherit            = true
  replication        = false
  superuser          = false
}

# Create dev database
resource "postgresql_database" "banterbus_dev" {
  provider = postgresql.homelab
  name  = "banterbus_dev"
  owner = postgresql_role.banterbus_dev.name
}

# Create prod database
resource "postgresql_database" "banterbus_prod" {
  provider = postgresql.homelab
  name  = "banterbus_prod"
  owner = postgresql_role.banterbus_prod.name
}

# Grant permissions for dev environment
resource "postgresql_grant" "banterbus_dev_connect" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_dev.name
  role        = postgresql_role.banterbus_dev.name
  object_type = "database"
  privileges  = ["CONNECT"]
}

resource "postgresql_grant" "banterbus_dev_schema" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_dev.name
  role        = postgresql_role.banterbus_dev.name
  schema      = "public"
  object_type = "schema"
  privileges  = ["USAGE", "CREATE"]
}

resource "postgresql_grant" "banterbus_dev_tables" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_dev.name
  role        = postgresql_role.banterbus_dev.name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
}

resource "postgresql_grant" "banterbus_dev_sequences" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_dev.name
  role        = postgresql_role.banterbus_dev.name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["USAGE", "SELECT", "UPDATE"]
}

# Grant permissions for prod environment
resource "postgresql_grant" "banterbus_prod_connect" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_prod.name
  role        = postgresql_role.banterbus_prod.name
  object_type = "database"
  privileges  = ["CONNECT"]
}

resource "postgresql_grant" "banterbus_prod_schema" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_prod.name
  role        = postgresql_role.banterbus_prod.name
  schema      = "public"
  object_type = "schema"
  privileges  = ["USAGE", "CREATE"]
}

resource "postgresql_grant" "banterbus_prod_tables" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_prod.name
  role        = postgresql_role.banterbus_prod.name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
}

resource "postgresql_grant" "banterbus_prod_sequences" {
  provider    = postgresql.homelab
  database    = postgresql_database.banterbus_prod.name
  role        = postgresql_role.banterbus_prod.name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["USAGE", "SELECT", "UPDATE"]
}

# Output connection strings
output "postgres_dev_connection_string" {
  value = "postgres://${postgresql_role.banterbus_dev.name}:${random_password.banterbus_dev.result}@${var.postgres_host}:${var.postgres_port}/${postgresql_database.banterbus_dev.name}?sslmode=disable"
  sensitive = true
}

output "postgres_prod_connection_string" {
  value = "postgres://${postgresql_role.banterbus_prod.name}:${random_password.banterbus_prod.result}@${var.postgres_host}:${var.postgres_port}/${postgresql_database.banterbus_prod.name}?sslmode=disable"
  sensitive = true
}