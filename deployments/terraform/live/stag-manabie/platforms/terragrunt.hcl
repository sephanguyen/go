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
    lms = {
      enabled          = true
      name             = "manabie-lms"
      random_suffix    = true
      region           = "asia-southeast1"
      zone             = "asia-southeast1-b"
      database_version = "POSTGRES_14"
      tier             = "db-custom-1-3840" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 50
      disk_autoresize  = false
      database_flags = [
        # TODO: use common database_flags config
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
        {
          name  = "pgaudit.role"
          value = "auditor"
        },
        {
          name  = "max_wal_senders"
          value = "10"
        },
        {
          name  = "max_replication_slots"
          value = "10"
        },
        {
          name  = "idle_in_transaction_session_timeout"
          value = "300000"
        },
        {
          name  = "log_min_duration_statement"
          value = "300000" // 300 seconds
        },
      ]

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia-southeast1"
      backup_start_time              = "22:00"
      point_in_time_recovery_enabled = false
      transaction_log_retention_days = 7
      retained_backups               = 3
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }

    common = {
      enabled          = true
      name             = "manabie-common"
      random_suffix    = true
      region           = "asia-southeast1"
      zone             = "asia-southeast1-b"
      database_version = "POSTGRES_14"
      tier             = "db-custom-4-15360" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 50
      disk_autoresize  = false
      database_flags = [
        # TODO: use common database_flags config
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
        {
          name  = "pgaudit.role"
          value = "auditor"
        },
        {
          name  = "max_wal_senders"
          value = "30"
        },
        {
          name  = "max_replication_slots"
          value = "30"
        },
        {
          name  = "max_connections"
          value = "1000"
        },
        {
          name  = "idle_in_transaction_session_timeout"
          value = "600000"
        },
        {
          name  = "log_min_duration_statement"
          value = "300000" // 300 seconds
        },
      ]

      insights_config = merge(
        include.env.locals.insights_config,
        {
          query_string_length = 2048,
        },
      )

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia-southeast1"
      backup_start_time              = "22:00"
      point_in_time_recovery_enabled = false
      transaction_log_retention_days = 7
      retained_backups               = 3
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }

    auth = {
      enabled          = true
      name             = "manabie-auth"
      random_suffix    = true
      region           = "asia-southeast1"
      zone             = "asia-southeast1-b"
      database_version = "POSTGRES_14"
      tier             = "db-custom-1-3840" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 50
      disk_autoresize  = false
      database_flags = [
        # TODO: use common database_flags config
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
        {
          name  = "pgaudit.role"
          value = "auditor"
        },
        {
          name  = "max_wal_senders"
          value = "10"
        },
        {
          name  = "max_replication_slots"
          value = "10"
        },
        {
          name  = "max_connections"
          value = "200"
        },
        {
          name  = "idle_in_transaction_session_timeout"
          value = "600000"
        },
        {
          name  = "log_min_duration_statement"
          value = "300000" // 300 seconds
        },
      ]

      insights_config = merge(
        include.env.locals.insights_config,
        {
          query_string_length = 2048,
        },
      )

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia-southeast1"
      backup_start_time              = "22:00"
      point_in_time_recovery_enabled = false
      transaction_log_retention_days = 7
      retained_backups               = 3
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }

  backend_bucket = {
    enabled       = true
    bucket_name   = "stag-manabie-backend"
    location      = "asia-southeast1"
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

  gke = {
    enabled        = false
    cluster_name   = "staging"
    region         = "asia-southeast1"
    regional       = true
    zones          = ["asia-southeast1-a"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.21.5-gke.1302"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = "gke-staging-pods-2af22c28"
    ip_range_services = "gke-staging-services-2af22c28"

    create_service_account = false
    service_account        = "tf-gke-staging-p642@staging-manabie-online.iam.gserviceaccount.com"

    gce_pd_csi_driver = true

    cluster_autoscaling = {
      enabled             = false
      autoscaling_profile = "OPTIMIZE_UTILIZATION"
      min_cpu_cores       = 0
      max_cpu_cores       = 0
      min_memory_gb       = 0
      max_memory_gb       = 0
      gpu_resources       = []
    }

    monitoring_config = {
      enable_components = [
        "APISERVER",
        "CONTROLLER_MANAGER",
        "SCHEDULER",
        "SYSTEM_COMPONENTS",
      ]
    }

    node_pools = [
      # runners pools
      {
        name         = "n2d-standard-4-runners-spot"
        machine_type = "n2d-standard-4"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 20
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 30
      },
      {
        name         = "n2d-standard-4-runners-on-demand"
        machine_type = "n2d-standard-4"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 20
        image_type   = "COS_CONTAINERD"
        spot         = false
        disk_size_gb = 30
      },
      {
        name         = "e2-highmem-2-runners-on-demand"
        machine_type = "e2-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 20
        image_type   = "COS_CONTAINERD"
        spot         = false
        disk_size_gb = 60
      },
      {
        name         = "t2d-standard-1-runners-on-demand"
        machine_type = "t2d-standard-1"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = false
        disk_size_gb = 50
      },
      {
        name         = "t2d-standard-1-runners-spot"
        machine_type = "t2d-standard-1"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 20
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },
      {
        name         = "c2d-highcpu-2-runners-on-demand"
        machine_type = "c2d-highcpu-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 10
        image_type   = "COS_CONTAINERD"
        spot         = false
        disk_size_gb = 50
      },
      {
        name         = "n2d-standard-8-runners-heavy-spot"
        machine_type = "n2d-standard-8"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 100
      },
      {
        name         = "runner-8-13-spot"
        machine_type = "n2d-custom-8-16384"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },
      {
        name         = "c2d-highcpu-4-runners-spot"
        machine_type = "c2d-highcpu-4"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },
      {
        name         = "n2d-highmem-2-runners-spot"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },

      # services pools
      {
        name         = "n2d-highmem-2-on-demand"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 3
        max_count    = 4
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "n2d-standard-2-spot"
        machine_type = "n2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 2
        max_count    = 6
        image_type   = "COS_CONTAINERD"
        spot         = true
      },
    ]

    node_pools_labels = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-8-runners-heavy-spot = {
        spot-ci-heavy = true
      }
    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }
    }

    node_pools_taints = {
      all = []

      "e2-standard-2-runners-on-demand" = [
        {
          key    = "e2-standard-2-runners-on-demand"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "n2d-standard-4-runners-spot" = [
        {
          key    = "n2d-standard-4-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "n2d-standard-4-runners-on-demand" = [
        {
          key    = "n2d-standard-4-runners-on-demand"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "e2-highmem-2-runners-on-demand" = [
        {
          key    = "e2-highmem-2-runners-on-demand"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "t2d-standard-1-runners-on-demand" = [
        {
          key    = "t2d-standard-1-runners-on-demand"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "c2d-highcpu-2-runners-on-demand" = [
        {
          key    = "c2d-highcpu-2-runners-on-demand"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "t2d-standard-1-runners-spot" = [
        {
          key    = "t2d-standard-1-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "n2d-standard-8-runners-heavy-spot" = [
        {
          key    = "spot-ci-heavy"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]
      "runner-8-13-spot" = [
        {
          key    = "runner-8-13"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "c2d-highcpu-4-runners-spot" = [
        {
          key    = "c2d-highcpu-4-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "n2d-highmem-2-runners-spot" = [
        {
          key    = "n2d-highmem-2-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      n2d-standard-2-spot = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]
    }

    node_pools_resource_labels = null

    node_pools_tags = {
      default_values = [false, false]
    }

    maintenance_start_time = "1970-01-01T18:00:00Z"
    maintenance_end_time   = "1970-01-01T22:00:00Z"
    maintenance_recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"
  }

  gke_rbac = {
    enabled = false
    policies = [
      {
        kind       = "Group"
        group      = "tech-func-backend@manabie.com"
        role_kind  = "ClusterRole"
        role_name  = "view"
        namespaces = []
      },
    ]
  }

  gke_enable_platforms_monitoring = true

  kms = {
    enabled  = true
    location = "global"
    keyring  = "deployments"
  }

  bigquery = {
    enabled                    = true
    delete_contents_on_destroy = false
    dataset_id                 = "manabie_dataset"
    dataset_name               = "manabie_dataset"
    description                = "BigQuery dataset for staging"
    location                   = "asia-southeast1"
    dataset_labels = {
      env      = "stag"
      org      = "manabie"
      billable = "false"
    }
  }

  import_map_deployer_bucket = {
    "staging-manabie-online" : {
      project_id    = "staging-manabie-online"
      bucket_name   = "import-map-deployer-staging"
      location      = "asia"
      storage_class = "STANDARD"

      cors = {
        origins = ["*"]
        max_age_seconds = 3600
        response_header = ["*"]
        methods         = ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"]
      }

      lifecycle_rule = [
        {
          condition = {
            age = 7
          }
          action = {
            type = "Delete"
          }
        }
      ]
    },

    "uat-manabie" : {
      project_id    = "uat-manabie"
      bucket_name   = "import-map-deployer-uat"
      location      = "asia-southeast1"
      storage_class = "STANDARD"

      cors = {
        origins = ["*"]
        max_age_seconds = 3600
        response_header = ["*"]
        methods         = ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"]
      }

      lifecycle_rule = [
        {
          condition = {
            age = 30
          }
          action = {
            type = "Delete"
          }
        }
      ]
    }

    "pre-prod" : {
      project_id    = "uat-manabie"
      bucket_name   = "import-map-deployer-preproduction"
      location      = "asia-southeast1"
      storage_class = "STANDARD"

      cors = {
        origins = ["*"]
        max_age_seconds = 3600
        response_header = ["*"]
        methods         = ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"]
      }

      lifecycle_rule = [
        {
          condition = {
            age = 30
          }
          action = {
            type = "Delete"
          }
        }
      ]
    }
  }
}
