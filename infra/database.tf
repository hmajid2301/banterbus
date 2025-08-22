resource "postgresql_database" "banterbus_db" {
  name  = var.postgres_database
  owner = var.postgres_username
}

resource "postgresql_role" "banterbus_user" {
  name     = "${var.postgres_database}_user"
  login    = true
  password = var.postgres_password
}

resource "postgresql_grant" "banterbus_db_connect" {
  database    = postgresql_database.banterbus_db.name
  role        = postgresql_role.banterbus_user.name
  object_type = "database"
  privileges  = ["CONNECT"]
}

resource "postgresql_grant" "banterbus_schema_usage" {
  database    = postgresql_database.banterbus_db.name
  role        = postgresql_role.banterbus_user.name
  schema      = "public"
  object_type = "schema"
  privileges  = ["USAGE", "CREATE"]
}

resource "postgresql_grant" "banterbus_tables" {
  database    = postgresql_database.banterbus_db.name
  role        = postgresql_role.banterbus_user.name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
}

resource "postgresql_grant" "banterbus_sequences" {
  database    = postgresql_database.banterbus_db.name
  role        = postgresql_role.banterbus_user.name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["USAGE", "SELECT", "UPDATE"]
}

output "postgres_connection_string" {
  value = "postgres://${postgresql_role.banterbus_user.name}:${var.postgres_password}@${var.postgres_host}:${var.postgres_port}/${postgresql_database.banterbus_db.name}?sslmode=require"
  sensitive = true
}