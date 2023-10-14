locals {
  hasura_metadata_databases = [
    for db in var.postgresql.databases : db if length(regexall("hasura_metadata$", db)) > 0
  ]
  postgresql_schemas = distinct(flatten([
    for user_permission in var.postgresql_user_permissions :
    {
      "database" : user_permission.database
      "schema" : user_permission.schema,
    }
    if user_permission.schema != "" && user_permission.schema != "public"
  ]))
}

data "google_sql_database_instance" "instance" {
  name    = var.postgresql.instance
  project = var.postgresql.project_id
}

resource "google_sql_database" "databases" {
  for_each = toset(var.postgresql.databases)

  project  = var.postgresql.project_id
  name     = each.value
  instance = data.google_sql_database_instance.instance.name
}

resource "postgresql_schema" "schemas" {
  for_each = length(local.postgresql_schemas) > 0 ? {
    for s in local.postgresql_schemas :
    "${s.database}.${s.schema}" => s
    if s.schema != "public"
  } : {}

  name     = each.value.schema
  database = each.value.database
  owner    = "postgres"
}

# See: https://hasura.io/docs/latest/graphql/core/deployment/postgres-requirements/#2-a-single-role-to-manage-metadata-and-user-objects-in-the-same-database
resource "postgresql_extension" "hasura_pgcrypto" {
  for_each = toset(local.hasura_metadata_databases)

  name     = "pgcrypto"
  database = each.value

  depends_on = [google_sql_database.databases]
}

resource "postgresql_extension" "pgaudit" {
  for_each = var.pgaudit_enabled ? toset(var.postgresql.databases) : []

  name     = "pgaudit"
  database = each.value

  depends_on = [google_sql_database.databases]
}

resource "random_password" "postgresql_users" {
  for_each = toset([
    for user in var.postgresql.users : user
    if length(regexall("@.+\\.iam$", user)) == 0
    # don't need to generate password for cloud IAM service account user
  ])

  length  = 16
  special = true
}

resource "google_sql_user" "users" {
  for_each = toset(var.postgresql.users)

  project  = var.postgresql.project_id
  type     = length(regexall("@.+\\.iam$", each.value)) == 0 ? "BUILT_IN" : "CLOUD_IAM_SERVICE_ACCOUNT"
  name     = each.value
  instance = data.google_sql_database_instance.instance.name
  password = length(regexall("@.+\\.iam$", each.value)) == 0 ? random_password.postgresql_users[each.value].result : null
}

# These are the roles we need to grant to atlantis, so that in turn, atlantis has
# the permissions to grant objects created by these roles to everyone else.
#
# This includes:
#  - stag-bob-m@staging-manabie-online.iam and similar accounts
#  - stag-bob-h@staging-manabie-online.iam and similar accounts
#  - stag-unleash@staging-manabie-online.iam and similar accounts (unleash migrates and serves using the same account)
resource "postgresql_grant_role" "migration_to_atlantis" {
  for_each = toset([
    for user in var.postgresql.users : user
    if length(regexall("hasura$", user)) > 0 || length(regexall("-h@.+\\.iam", user)) > 0 || length(regexall("-m@.+\\.iam", user)) > 0
    || length(regexall("unleash", user)) > 0 || length(regexall("-kafka-connect@.+\\.iam", user)) > 0
  ])

  role       = "atlantis@student-coach-e1e95.iam"
  grant_role = each.value

  depends_on = [google_sql_user.users]
}

# Grant ad-hoc full permissions from all migration accounts, so that
# it can freely execute DDLs.
resource "postgresql_grant_role" "migration_to_adhoc" {
  for_each = toset([
    for user in var.postgresql.users : user
    if length(regexall("-m@.+\\.iam", user)) > 0 && var.adhoc.grant_enabled
  ])

  role       = var.adhoc.dbuser
  grant_role = each.value

  depends_on = [google_sql_user.users]
}

resource "postgresql_grant" "user_permissions" {
  for_each = length(var.postgresql.users) > 0 ? {
    for p in var.postgresql_user_permissions :
    "${p.database}.${p.user}.${p.schema}.${p.object_type}" => p
  } : {}

  database    = google_sql_database.databases[each.value.database].name
  role        = google_sql_user.users[each.value.user].name
  schema      = each.value.schema
  object_type = each.value.object_type
  privileges  = each.value.privileges
  objects     = try(each.value.objects, [])

  depends_on = [
    postgresql_grant_role.migration_to_atlantis,
    postgresql_schema.schemas
  ]
}

resource "postgresql_default_privileges" "user_default_permissions" {
  for_each = length(var.postgresql.users) > 0 ? {
    for p in var.postgresql_user_permissions :
    # If p.owner is "postgres", we don't insert it into the key string
    # since p.owner was added much later and we wanted to avoid re-creating all the resources
    "${p.database}.${p.user}${p.owner == "postgres" ? "" : format(".%s", p.owner)}.${p.schema}.${p.object_type}" => p
    # if the objects to be granted exists, that means the privileges only apply for
    # some objects, so we don't need to set default privileges for that object_type
    if length(coalesce(p.objects, [])) == 0 &&
    p.object_type != "database" && p.object_type != "schema" # default privilege doesn't work with database & schema object_type
  } : {}

  database    = google_sql_database.databases[each.value.database].name
  role        = google_sql_user.users[each.value.user].name
  owner       = each.value.owner
  schema      = each.value.schema
  object_type = each.value.object_type
  privileges  = each.value.privileges

  depends_on = [
    postgresql_grant_role.migration_to_atlantis,
    postgresql_schema.schemas
  ]
}

resource "null_resource" "grant_bypass_rls_roles" {
  for_each = toset(var.postgresql_bypass_rls_roles)

  triggers = {
    role = each.value
  }

  provisioner "local-exec" {
    command = <<EOT
      psql \
        -h 127.0.0.1 \
        -p ${var.postgresql_instance_port[var.postgresql.instance]} \
        -U "atlantis@student-coach-e1e95.iam" \
        -d postgres \
        -c 'ALTER ROLE "${each.value}" BYPASSRLS'
    EOT
  }

  depends_on = [
    # depends on google_sql_user resource
    # since the role must exist in the DB first
    # before we can alter that role
    google_sql_user.users,
  ]
}

resource "null_resource" "grant_connection_limit_roles" {
  for_each = toset(var.postgresql.users)

  triggers = {
    role = each.value
  }

  provisioner "local-exec" {
    command = <<EOT
      psql \
        -h 127.0.0.1 \
        -p ${var.postgresql_instance_port[var.postgresql.instance]} \
        -U "atlantis@student-coach-e1e95.iam" \
        -d postgres \
        -c 'ALTER ROLE "${each.value}" CONNECTION LIMIT 150'
    EOT
  }

  depends_on = [
    # depends on google_sql_user resource
    # since the role must exist in the DB first
    # before we can alter that role
    google_sql_user.users,
  ]
}

resource "null_resource" "grant_replication_roles" {
  for_each = toset(var.postgresql_replication_roles)

  triggers = {
    role = each.value
  }

  provisioner "local-exec" {
    command = <<EOT
      psql \
        -h 127.0.0.1 \
        -p ${var.postgresql_instance_port[var.postgresql.instance]} \
        -U "atlantis@student-coach-e1e95.iam" \
        -d postgres \
        -c 'ALTER ROLE "${each.value}" REPLICATION'
    EOT
  }

  depends_on = [
    # depends on google_sql_user resource
    # since the role must exist in the DB first
    # before we can alter that role
    google_sql_user.users,
  ]
}

resource "null_resource" "grant_statement_timeout_roles" {
  for_each = length(var.postgresql.users) > 0 ? {
    for p in var.postgresql_statement_timeout :
    "${p.user}.${p.statement_timeout}" => p
  } : {}

  triggers = {
    role = each.value.user
  }

  provisioner "local-exec" {
    command = <<EOT
      psql \
        -h 127.0.0.1 \
        -p ${var.postgresql_instance_port[var.postgresql.instance]} \
        -U "atlantis@student-coach-e1e95.iam" \
        -d postgres \
        -c 'ALTER ROLE "${each.value.user}" SET statement_timeout = "${each.value.statement_timeout}"'
    EOT
  }

  depends_on = [
    # depends on google_sql_user resource
    # since the role must exist in the DB first
    # before we can alter that role
    google_sql_user.users,
  ]
}
