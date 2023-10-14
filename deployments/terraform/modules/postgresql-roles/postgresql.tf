locals {
  member_function_level = flatten([
    for function, levels in var.member_by_access_level : [
      for level, members in levels : [
        for member in members : {
          member   = member
          function = function
          level    = level
        }
      ]
    ]
  ])

  postgresql_read_write_users = toset(distinct([
    for m in local.member_function_level : m.member
    if try(var.role_by_access_level[m.function][m.level][var.env].can_write_databases, false)
  ]))

  // if a member already exists in the read_write_group,
  // exclude that member from this read_only_group.
  postgresql_read_only_users = setsubtract(
    toset(distinct([
      for m in local.member_function_level : m.member
      if try(var.role_by_access_level[m.function][m.level][var.env].can_read_databases, false)
    ])),
    local.postgresql_read_write_users,
  )
}

resource "postgresql_role" "read_only_role" {
  name = "read_only_role"
}

resource "postgresql_role" "read_write_role" {
  name = "read_write_role"
}

resource "postgresql_role" "bypass_rls_role" {
  name                      = "bypass_rls_role"
  bypass_row_level_security = true

  lifecycle {
    ignore_changes = [
      roles, # ignore roles changes because there maybe other projects grant roles to this role
    ]
  }
}

resource "google_sql_user" "postgresql_read_only_users" {
  for_each = local.postgresql_read_only_users

  project  = var.postgresql_project_id
  name     = each.value
  instance = var.postgresql_instance
  type     = "CLOUD_IAM_USER"
}

resource "google_sql_user" "postgresql_read_write_users" {
  for_each = local.postgresql_read_write_users

  project  = var.postgresql_project_id
  name     = each.value
  instance = var.postgresql_instance
  type     = "CLOUD_IAM_USER"
}

output "postgresql_read_only_users" {
  value = local.postgresql_read_only_users
}

output "postgresql_read_write_users" {
  value = local.postgresql_read_write_users
}
