include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/platforms"
}

dependency "vpc" {
  // using the same vpc in prod-tokyo
  config_path = "..//vpc"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/platforms.hcl"
  expose = true
}

inputs = {
  project_id = "student-coach-e1e95"

  postgresql = {
    analytics = {
      enabled          = true
      name             = "analytics"
      random_suffix    = false
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_14"
      tier             = "db-g1-small"
      disk_size        = 15
      disk_autoresize  = true
      database_flags   = include.env.locals.database_flags

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia-northeast1"
      backup_start_time              = "22:00"
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
      retained_backups               = 7
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }
}
