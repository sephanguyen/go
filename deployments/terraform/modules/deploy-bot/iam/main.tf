/**
* This module grants the deploy bot (specified from input) access to the specified GCP Cloud SQL database.
*/
locals {
  roleList = [
    {
      id          = "dorp.deploybot.backupDatabase"
      title       = "Backup Database"
      description = "Preproduction Deploy Bot's role to backup production database"

      base_roles = []
      permissions = [
        "cloudsql.backupRuns.create",
        "cloudsql.backupRuns.get",
        "cloudsql.backupRuns.list",
        "cloudsql.instances.get",
        "cloudsql.instances.list",
      ]
    },
    {
      id          = "dorp.deploybot.restoreDatabaseBackup"
      title       = "Restore Database Backup"
      description = "Preproduction Deploy Bot's role to restore database backups"

      base_roles = []
      permissions = [
        "cloudsql.instances.get",
        "cloudsql.instances.list",
        "cloudsql.instances.restoreBackup",
      ]
    },
  ]

  roles = {
    for r in local.roleList : r.id => r
  }
}

module "custom-roles" {
  source  = "terraform-google-modules/iam/google//modules/custom_role_iam"
  version = "7.4.1"

  for_each = local.roles

  target_level = "project"
  target_id    = var.project_id

  role_id     = each.value.id
  title       = each.value.title
  description = each.value.description

  base_roles  = each.value.base_roles
  permissions = each.value.permissions

  members = []
}

# This resource allows dorp_deploy_bot to backup production databases.
resource "google_project_iam_member" "dorp_deploy_bot_backup_database_role" {
  for_each = {
    for db in var.production_databases : format("%s/%s", db.project_id, db.instance_id) => db
  }

  project = var.project_id
  role    = "projects/${var.project_id}/roles/dorp.deploybot.backupDatabase"
  member  = "serviceAccount:${var.dorp_deploy_bot_service_account.email}"

  condition {
    title       = "dorp_deploy_bot_backup_database"
    description = "Allowing Deploy Bot to backup production database"
    expression  = "resource.name.startsWith('projects/${each.value.project_id}/instances/${each.value.instance_id}') && resource.service == 'sqladmin.googleapis.com'"
  }
}

# This resource allows dorp_deploy_bot to restore backup for preproduction databases.
resource "google_project_iam_member" "dorp_deploy_bot_restore_database_backups_role" {
  for_each = {
    for db in var.preproduction_databases : format("%s/%s", db.project_id, db.instance_id) => db
  }

  project = var.project_id
  role    = "projects/${var.project_id}/roles/dorp.deploybot.restoreDatabaseBackup"
  member = "serviceAccount:${var.dorp_deploy_bot_service_account.email}"

  condition {
    title       = "dorp_deploy_bot_restore_database_backups"
    description = "Allowing Deploy Bot to restore database backups"
    expression  = "resource.name.startsWith('projects/${each.value.project_id}/instances/${each.value.instance_id}') && resource.service == 'sqladmin.googleapis.com'"
  }
}

# This resource allows dorp_deplot_bot to connect to GKE cluster.
resource "google_project_iam_member" "dorp_deploy_bot_container_viewer" {
  project = var.project_id
  role    = "roles/container.viewer"
  member  = "serviceAccount:${var.dorp_deploy_bot_service_account.email}"
}
