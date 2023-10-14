locals {
  migrate_by_service_accounts = (var.env == "stag" || var.env == "uat")
}

resource "postgresql_grant" "read_only_role_select_tables" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_only_role_name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT"]
}

resource "postgresql_grant" "read_only_role_select_sequences" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_only_role_name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["SELECT"]
}

resource "postgresql_default_privileges" "read_only_role_select_tables" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_only_role_name
  schema      = "public"
  owner       = local.migrate_by_service_accounts ? "atlantis@student-coach-e1e95.iam" : "postgres"
  object_type = "table"
  privileges  = ["SELECT"]
}

resource "postgresql_default_privileges" "read_only_role_select_sequences" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_only_role_name
  schema      = "public"
  owner       = local.migrate_by_service_accounts ? "atlantis@student-coach-e1e95.iam" : "postgres"
  object_type = "sequence"
  privileges  = ["SELECT"]
}

resource "postgresql_grant" "read_write_role_write_tables" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_write_role_name
  schema      = "public"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE"]
}

resource "postgresql_grant" "read_write_role_write_sequences" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_write_role_name
  schema      = "public"
  object_type = "sequence"
  privileges  = ["SELECT", "USAGE", "UPDATE"]
}

resource "postgresql_default_privileges" "read_write_role_write_tables" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_write_role_name
  schema      = "public"
  owner       = local.migrate_by_service_accounts ? "atlantis@student-coach-e1e95.iam" : "postgres"
  object_type = "table"
  privileges  = ["SELECT", "INSERT", "UPDATE"]
}

resource "postgresql_default_privileges" "read_write_role_write_sequences" {
  for_each = toset(var.postgresql_databases)

  database    = each.value
  role        = var.postgresql_read_write_role_name
  schema      = "public"
  owner       = local.migrate_by_service_accounts ? "atlantis@student-coach-e1e95.iam" : "postgres"
  object_type = "sequence"
  privileges  = ["SELECT", "USAGE", "UPDATE"]
}
