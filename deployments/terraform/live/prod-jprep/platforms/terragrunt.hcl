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
    jprep = {
      enabled          = true
      name             = "jprep-6a98"
      random_suffix    = false
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_11"
      tier             = "db-custom-4-24576"
      disk_size        = 20
      disk_autoresize  = true
      database_flags = [
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
          name  = "cloudsql.enable_pgaudit"
          value = "on"
        },
        {
          name  = "pgaudit.log"
          value = "ddl"
        }
      ]

      insights_config = include.env.locals.insights_config

      deletion_protection = true

      maintenance_window_day          = 7
      maintenance_window_hour         = 21
      maintenance_window_update_track = "stable"

      backup_location                = "asia"
      backup_start_time              = "20:00"
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 3
      retained_backups               = 7
      retention_unit                 = "COUNT"

      private_network = dependency.vpc.outputs.network_self_link
    }
  }

  backend_bucket = {
    enabled       = true
    bucket_name   = "jprep-backend"
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

  gke_enable_resources_monitoring = true
  gke_enable_platforms_monitoring = true

  gke = {
    enabled        = true
    cluster_name   = "production"
    region         = "asia-northeast1"
    regional       = false
    zones          = ["asia-northeast1-c"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.19.14-gke.1900"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = "gke-production-pods-a761e181"
    ip_range_services = "gke-production-services-a761e181"

    create_service_account = false
    service_account        = "tf-gke-production-9kqn@live-manabie.iam.gserviceaccount.com"

    gce_pd_csi_driver = false

    cluster_autoscaling = {
      enabled             = false
      autoscaling_profile = "BALANCED"
      max_cpu_cores       = 0
      min_cpu_cores       = 0
      max_memory_gb       = 0
      min_memory_gb       = 0
      gpu_resources       = []
    }

    node_pools = [
      {
        name         = "n2d-custom-2-16"
        machine_type = "n2d-custom-2-16384"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "n2d-standard-2"
        machine_type = "n2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "e2-medium"
        machine_type = "e2-medium"
        autoscaling  = true
        node_count   = null
        min_count    = 2
        max_count    = 5
        image_type   = "COS_CONTAINERD"
        spot         = true
      },
      {
        name         = "n2d-standard-2-spot"
        machine_type = "n2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 3
        image_type   = "COS_CONTAINERD"
        spot         = true
      }
    ]

    node_pools_labels = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-4 = {
        cluster_name = "production"
        node_pool    = "n2d-standard-4"
      }

      n2d-standard-2 = {
        cluster_name = "production"
        node_pool    = "n2d-standard-2"
      }

      e2-medium = {
        cluster_name = "production"
        node_pool    = "e2-medium"
      }

      n2d-standard-2-spot = {
        cluster_name = "production"
        node_pool    = "n2d-standard-2"
      }

    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-4 = {
        cluster_name = "production"
        node_pool    = "n2d-standard-4"
      }
    }

    node_pools_taints = {
      pool-monitoring = [
        {
          key    = "monitoring"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      e2-medium = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
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

      n2d-standard-4 = [
        "gke-production",
        "gke-production-n2d-standard-4",
      ]
    }

    maintenance_start_time = "1970-01-01T18:00:00Z"
    maintenance_end_time   = "1970-01-01T22:00:00Z"
    maintenance_recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"
  }

  gke_rbac = {
    enabled = true
    policies = [
      {
        kind       = "Group"
        group      = "tech-func-backend@manabie.com"
        role_kind  = "ClusterRole"
        role_name  = "view"
        namespaces = []
      },
      {
        kind       = "Group"
        group      = "tech-func-platform@manabie.com"
        role_kind  = "ClusterRole"
        role_name  = "view"
        namespaces = []
      },
    ]
  }

  kms = {
    enabled  = true
    location = "global"
    keyring  = "deployments"
  }
}
