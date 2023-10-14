resource "google_gke_backup_backup_plan" "backup_plan" {
  for_each = {
    for plan in var.gke_backup_plan : plan.name => plan
  }

  project  = each.value.project
  cluster  = each.value.cluster
  name     = each.value.name
  location = each.value.location

  retention_policy {
    backup_delete_lock_days = each.value.retention_policy.backup_delete_lock_days
    backup_retain_days      = each.value.retention_policy.backup_retain_days
  }

  backup_schedule {
    cron_schedule = each.value.cron_schedule
  }

  backup_config {
    include_volume_data = each.value.backup_config.include_volume_data
    include_secrets     = each.value.backup_config.include_secrets

    selected_applications {

      dynamic "namespaced_names" {
        for_each = each.value.backup_config.selected_applications
        content {
          namespace = namespaced_names.value["namespace"]
          name      = namespaced_names.value["name"]
        }
      }
    }
  }
}
