include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/platforms"
}

dependency "vpc" {
  config_path = "../vpc"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/platforms.hcl"
  expose = true
}

inputs = {
  postgresql = {
    synersia = {
      enabled          = false
      name             = "synersia-228d"
      random_suffix    = false
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_11"
      tier             = "db-custom-2-7680"
      disk_size        = 20
      disk_autoresize  = true
      database_flags   = include.env.locals.database_flags

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia"
      backup_start_time              = "22:00"
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
      retained_backups               = 7
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }

  gke_enable_platforms_monitoring = false

  backend_bucket = {
    enabled       = true
    bucket_name   = "synersia-backend"
    location      = "asia-northeast1"
    storage_class = "STANDARD"
    cors = [
      {
        max_age_seconds = 3600
        method = [
          "GET",
          "HEAD",
          "POST",
          "PUT",
          "DELETE",
          "OPTIONS",
        ]
        origin = [
          "*",
        ]
        response_header = ["*"]
      }
    ]

    versioning_enabled          = true
    uniform_bucket_level_access = false
  }
}
