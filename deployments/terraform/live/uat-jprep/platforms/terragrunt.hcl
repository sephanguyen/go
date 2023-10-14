include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/platforms"
}

dependency "vpc" {
  config_path = "../../stag-manabie/vpc"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/platforms.hcl"
  expose = true
}

inputs = {
  postgresql = {
    jprep = {
      enabled          = true
      name             = "jprep-uat"
      random_suffix    = false
      region           = "asia-southeast1"
      zone             = "asia-southeast1-a"
      database_version = "POSTGRES_14"
      tier             = "db-custom-2-7680"
      disk_size        = 30
      disk_autoresize  = false
      database_flags = [
        {
          name  = "autovacuum"
          value = "on"
        },
        {
          name  = "cloudsql.iam_authentication"
          value = "on"
        },
        {
          name  = "cloudsql.logical_decoding"
          value = "on"
        },
        {
          name  = "pgaudit.log"
          value = "ddl"
        },
        {
          name  = "cloudsql.enable_pgaudit"
          value = "on"
        },
        {
          name  = "max_connections"
          value = "600"
        },
        {
          name  = "log_min_duration_statement"
          value = "300000" // 300 seconds
        },
        {
          name  = "max_wal_senders"
          value = "30"
        },
        {
          name  = "max_replication_slots"
          value = "30"
        },
      ]

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 5
      maintenance_window_hour         = 19
      maintenance_window_update_track = "stable"

      backup_location                = "asia-southeast1"
      backup_start_time              = "00:00"
      point_in_time_recovery_enabled = false
      transaction_log_retention_days = 7
      retained_backups               = 3
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }

  postgresql_alerts = {
    memory_utilization = {
      # Higher than usual to prevent false positive. See https://manabie.atlassian.net/browse/LT-35431
      threshold_value = 0.95
    }
  }
}
