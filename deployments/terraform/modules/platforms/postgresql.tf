module "postgresql" {
  for_each = {
    for name, config in var.postgresql : name => config
    if config.enabled
  }

  source  = "GoogleCloudPlatform/sql-db/google//modules/postgresql"
  version = "14.0.1"

  project_id           = var.project_id
  name                 = each.value.name
  random_instance_name = each.value.random_suffix

  region           = each.value.region
  zone             = each.value.zone
  tier             = each.value.tier
  disk_size        = each.value.disk_size
  disk_autoresize  = each.value.disk_autoresize
  database_version = each.value.database_version

  enable_default_db = false

  deletion_protection         = each.value.deletion_protection
  deletion_protection_enabled = each.value.deletion_protection

  user_name = "postgres"

  database_flags = each.value.database_flags

  insights_config = each.value.insights_config

  maintenance_window_day          = each.value.maintenance_window_day
  maintenance_window_hour         = each.value.maintenance_window_hour
  maintenance_window_update_track = each.value.maintenance_window_update_track

  backup_configuration = {
    enabled                        = true
    location                       = each.value.backup_location
    start_time                     = each.value.backup_start_time
    point_in_time_recovery_enabled = each.value.point_in_time_recovery_enabled
    transaction_log_retention_days = each.value.transaction_log_retention_days
    retained_backups               = each.value.retained_backups
    retention_unit                 = each.value.retention_unit
  }

  ip_configuration = {
    ipv4_enabled        = true
    private_network     = each.value.private_network
    require_ssl         = false
    authorized_networks = each.value.authorized_networks
    allocated_ip_range  = null
  }
}

data "google_monitoring_notification_channel" "slack" {
  project      = var.project_id
  type         = "slack"
  display_name = "Monitoring channel"
}

resource "google_monitoring_alert_policy" "cloud_sql" {
  for_each = {
    for name, config in var.postgresql : name => config
    if config.enabled
  }

  project      = var.project_id
  combiner     = "OR"
  display_name = "[${module.postgresql[each.key].instance_name}] Cloud SQL resource utilization is too high"

  conditions {
    display_name = "CPU utilization is more than 80%"

    condition_threshold {
      filter = <<-EOT
        resource.type="cloudsql_database"
        resource.label.database_id="${var.project_id}:${module.postgresql[each.key].instance_name}"
        metric.type="cloudsql.googleapis.com/database/cpu/utilization"
      EOT

      duration        = "180s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.8

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  conditions {
    display_name = "Memory utilization is more than ${var.postgresql_alerts.memory_utilization.threshold_value * 100}%"

    condition_threshold {
      filter = <<-EOT
        resource.type="cloudsql_database"
        resource.label.database_id="${var.project_id}:${module.postgresql[each.key].instance_name}"
        metric.type="cloudsql.googleapis.com/database/memory/utilization"
      EOT

      duration        = var.postgresql_alerts.memory_utilization.duration
      comparison      = var.postgresql_alerts.memory_utilization.comparison
      threshold_value = var.postgresql_alerts.memory_utilization.threshold_value

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  conditions {
    display_name = "Wal size is increasing over max size"

    condition_threshold {
      filter = <<-EOT
        resource.type="cloudsql_database"
        resource.label.database_id="${var.project_id}:${module.postgresql[each.key].instance_name}"
        metric.type="cloudsql.googleapis.com/database/disk/bytes_used_by_data_type"
        metric.label.data_type = "Wal"
      EOT

      duration        = "180s"
      comparison      = "COMPARISON_GT"
      threshold_value = 4000000000 // 4Gi
      trigger {
        count = 1
      }
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

resource "google_monitoring_alert_policy" "cloud_sql_slow_queries" {
  for_each = {
    for name, config in var.postgresql : name => config
    if config.enabled
  }

  project      = var.project_id
  combiner     = "OR"
  display_name = "[${module.postgresql[each.key].instance_name}] Slow SQL queries"

  conditions {
    display_name = "High number of slow SQL queries"
    condition_matched_log {
      filter = <<-EOT
        resource.type="cloudsql_database"
        resource.labels.database_id="${var.project_id}:${module.postgresql[each.key].instance_name}"
        textPayload=~"duration:\s[0-9.]+"
      EOT
    }
  }

  alert_strategy {
    notification_rate_limit {
      period = "300s"
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}
