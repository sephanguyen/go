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
    if coalesce(try(var.role_by_access_level[m.function][m.level][var.env].can_write_databases, null), false)
  ]))

  // if a member already exists in the read_write_group,
  // exclude that member from this read_only_group.
  postgresql_read_only_users = setsubtract(
    toset(distinct([
      for m in local.member_function_level : m.member
      if coalesce(try(var.role_by_access_level[m.function][m.level][var.env].can_read_databases, null), false)
    ])),
    local.postgresql_read_write_users,
  )

  postgresql_pg_write_all_data_users = var.postgresql_use_predefined_roles ? local.postgresql_read_write_users : []
  postgresql_pg_read_all_data_users  = var.postgresql_use_predefined_roles ? setunion(local.postgresql_read_write_users, local.postgresql_read_only_users) : []

  postgresql_bypass_rls_users = flatten([
    for function, values in var.member_by_access_level : [
      for level, emails in values : emails
      if contains(var.access_level_to_grant_by_pass_rls, level)
    ]
  ])
}

resource "postgresql_grant_role" "pg_read_all_data_users" {
  for_each = local.postgresql_pg_read_all_data_users

  role       = each.value
  grant_role = "pg_read_all_data"
}

resource "postgresql_grant_role" "read_only_users" {
  for_each = local.postgresql_read_only_users

  role       = each.value
  grant_role = var.postgresql_read_only_role_name
}

resource "postgresql_grant_role" "pg_write_all_data_users" {
  for_each = local.postgresql_pg_write_all_data_users

  role       = each.value
  grant_role = "pg_write_all_data"
}

resource "postgresql_grant_role" "read_write_users" {
  for_each = local.postgresql_read_write_users

  role       = each.value
  grant_role = var.postgresql_read_write_role_name
}

output "postgresql_read_only_users" {
  value = local.postgresql_read_only_users
}

output "postgresql_read_write_users" {
  value = local.postgresql_read_write_users
}

# Steps to bypass rls:
#   1. connect to the database
#   2. SET ROLE bypass_rls_role;
#   3. query tables
# we can see that after running command at step 2) above, the
# current role is changed to bypass_rls_role, it's not the original
# log in user anymore.
# To make bypass_rls_role can query tables at step 3), we need
# to grant it into the read_only_role role (or read_write_role, if enabled)
# so it can have necessary privileges to query tables.
resource "postgresql_grant_role" "bypass_rls_read_only" {
  role       = "bypass_rls_role"
  grant_role = var.bypass_rls_role_write_privileges_enabled ? var.postgresql_read_write_role_name : var.postgresql_read_only_role_name
}

# By default, bypass_rls_role can only query tables, but not insert/update/delete.
resource "postgresql_grant_role" "bypass_rls_read_all_data" {
  count = var.postgresql_use_predefined_roles ? 1 : 0

  role       = "bypass_rls_role"
  grant_role = "pg_read_all_data"
}

# To enable insert/update/delete, we need to grant it into pg_write_all_data role.
# This is only enabled if var.bypass_rls_role_write_privileges_enabled is true.
resource "postgresql_grant_role" "bypass_rls_write_all_data" {
  count = var.postgresql_use_predefined_roles && var.bypass_rls_role_write_privileges_enabled ? 1 : 0

  role       = "bypass_rls_role"
  grant_role = "pg_write_all_data"
}

resource "postgresql_grant_role" "bypass_rls_users" {
  for_each = toset(local.postgresql_bypass_rls_users)

  role       = each.value
  grant_role = "bypass_rls_role"
}
