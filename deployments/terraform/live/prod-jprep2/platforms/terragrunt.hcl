include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/platforms"
}

# We use the same VPC with Tokyo project, because we want to deploy
# JPREP deployments to Tokyo GKE cluster.
dependency "vpc" {
  config_path = "../../prod-tokyo/vpc"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/platforms.hcl"
  expose = true
}

inputs = {
  postgresql = {
    jprep = {
      enabled          = true
      name             = "prod-jprep"
      random_suffix    = true
      region           = "asia-northeast1"
      zone             = "asia-northeast1-a"
      database_version = "POSTGRES_11"
      tier             = "db-custom-4-20480"
      disk_size        = 200
      disk_autoresize  = true
      # TODO: use database_flags common variable for this
      database_flags = [
        {
          name  = "log_min_duration_statement"
          value = "300000"
        },
        {
          name  = "log_checkpoints"
          value = "on"
        },
        {
          name  = "log_connections"
          value = "on"
        },
        {
          name  = "log_disconnections"
          value = "on"
        },
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
          name  = "wal_compression"
          value = "on"
        },
        {
          name  = "cloudsql.enable_pgaudit"
          value = "on"
        },
        {
          name  = "pgaudit.log"
          value = "ddl"
        },
      ]

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia-northeast1"
      backup_start_time              = "20:00"
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
      retained_backups               = 7
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }

  kms = {
    enabled  = true
    location = "asia-northeast1"
    keyring  = "jprep"
  }
}
