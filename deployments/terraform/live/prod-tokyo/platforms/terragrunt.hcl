include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/platforms"
}

dependency "vpc" {
  config_path = "..//vpc"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/platforms.hcl"
  expose = true
}

inputs = {
  project_id = "student-coach-e1e95"

  postgresql = {
    tokyo = {
      enabled          = true
      name             = "prod-tokyo"
      random_suffix    = false
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_13"
      tier             = "db-custom-8-16384"
      disk_size        = 15
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
          name  = "max_connections"
          value = "600"
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
          name  = "log_min_duration_statement"
          value = "300000"
        },
      ]

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

    lms = {
      enabled          = true
      name             = "prod-tokyo-lms"
      random_suffix    = true
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_14"
      tier             = "db-custom-2-6144" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 10
      disk_autoresize  = true
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
          name  = "log_min_duration_statement"
          value = "300000"
        },
        {
          name  = "max_connections"
          value = "400"
        },
      ]

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

    auth = {
      enabled          = true
      name             = "prod-tokyo-auth"
      random_suffix    = true
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_14"
      tier             = "db-g1-small" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 10
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
          name  = "log_min_duration_statement"
          value = "300000"
        },
        {
          name  = "max_connections"
          value = "200"
        },
      ]

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

    data-warehouse = {
      enabled          = true
      name             = "prod-tokyo-data-warehouse"
      random_suffix    = true
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_14"
      tier             = "db-f1-micro" // See: https://cloud.google.com/sql/docs/postgres/create-instance
      disk_size        = 10
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
          name  = "log_min_duration_statement"
          value = "300000"
        },
        {
          name  = "max_connections"
          value = "200"
        },
      ]

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

  backend_bucket = {
    enabled       = true
    bucket_name   = "prod-tokyo-backend"
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

    versioning_enabled          = false
    uniform_bucket_level_access = false
  }

  gke = {
    enabled        = true
    cluster_name   = "tokyo"
    region         = "asia-northeast1"
    regional       = true
    zones          = ["asia-northeast1-a"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.20.15-gke.3400"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = "gke-range-pods"
    ip_range_services = "gke-range-services"

    create_service_account = true
    service_account        = ""

    gce_pd_csi_driver   = true
    backup_agent_config = true

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
        name         = "pool-monitoring"
        machine_type = "n2d-custom-2-16384"
        autoscaling  = false
        node_count   = 1
        min_count    = null
        max_count    = null
        image_type   = "COS_CONTAINERD"
        spot         = false
        disk_type    = "pd-balanced"
        disk_size_gb = 50
      },
      {
        name         = "n2d-highmem-2-on-demand"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 6
        max_count    = 7
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "n2d-standard-2-on-demand"
        machine_type = "n2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 1
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "pool-backend-on-demand"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name           = "pool-kafka-on-demand"
        machine_type   = "n2d-highmem-2"
        autoscaling    = true
        node_count     = null
        min_count      = 0
        max_count      = 3
        image_type     = "COS_CONTAINERD"
        spot           = false
      },
      {
        name           = "n2d-highmem-2-spot-c"
        machine_type   = "n2d-highmem-2"
        autoscaling    = true
        node_count     = null
        min_count      = 2
        max_count      = 3
        image_type     = "COS_CONTAINERD"
        spot           = true
        node_locations = "asia-northeast1-c"
      },
      {
        name         = "t2d-standard-2-spot"
        machine_type = "t2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = true
      },
      {
        name         = "preproduction-n2d-highmem-2-spot"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 1
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

      n2d-highmem-2-on-demand = {
        cluster_name = "tokyo"
        node_pool    = "n2d-highmem-2-on-demand"
      }

      n2d-standard-2-on-demand = {
        cluster_name = "tokyo"
        node_pool    = "n2d-standard-2-on-demand"
      }

      pool-backend-on-demand = {
        backend-on-demand-node = "true"
      }

      pool-kafka-on-demand = {
        kafka = "true"
      }

      preproduction-n2d-highmem-2-spot = {
        cluster_name = "tokyo"
        node_pool    = "preproduction-n2d-highmem-2-spot"
      }

      "preproduction-n2d-highmem-2-spot" = {
        cluster_name = "tokyo"
        node_pool    = "preproduction-n2d-highmem-2-spot"
        environment  = "preproduction"
      }
    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-highmem-2-spot = {
        cluster_name = "tokyo"
        node_pool    = "n2d-highmem-2-spot"
      }
    }

    node_pools_taints = {
      pool-backend-on-demand = [
        {
          key    = "backend-on-demand-node"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      pool-kafka-on-demand = [
        {
          key    = "kafka"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      pool-monitoring = [
        {
          key    = "monitoring"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      n2d-highmem-2-spot-c = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      t2d-standard-2-spot = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      "preproduction-n2d-highmem-2-spot" = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
        {
          key    = "environment"
          value  = "preproduction"
          effect = "NO_SCHEDULE"
        },
      ]
    }

    node_pools_resource_labels = null

    node_pools_tags = {
      default_values = [false, false]

      n2d-highmem-2-spot = [
        "gke-tokyo",
        "gke-tokyo-n2d-highmem-2-spot",
      ]
    }

    maintenance_start_time = "1970-01-01T17:00:00Z"
    maintenance_end_time   = "1970-01-01T21:00:00Z"
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

  gke_enable_platforms_monitoring = true

  kubernetes_cluster_roles = [
    {
      # This "custom-admin" role has almost the same permissions as
      # predefined "admin" role, except that it doesn't have permission
      # to delete some resources, such as deployments, statefulsets, secrets...etc.
      # We can use this role to prevent accidental deletion of resources.
      name = "custom-admin"
      rules = [
        {
          api_groups = [""]
          resources  = ["pods/attach", "pods/exec", "pods/portforward", "pods/proxy", "pods/log", "secrets", "services/proxy"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = [""]
          resources  = ["pods", "pods/attach", "pods/exec", "pods/portforward", "pods/proxy"]
          verbs      = ["create", "delete", "deletecollection", "patch", "update"]
        },
        {
          api_groups = [""]
          resources  = ["configmaps", "endpoints", "persistentvolumeclaims", "replicationcontrollers", "replicationcontrollers/scale", "secrets", "serviceaccounts", "services", "services/proxy"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["apps"]
          resources  = ["daemonsets", "deployments", "deployments/rollback", "deployments/scale", "replicasets", "replicasets/scale", "statefulsets", "statefulsets/scale"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["autoscaling"]
          resources  = ["horizontalpodautoscalers"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["batch"]
          resources  = ["cronjobs", "jobs"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["extensions"]
          resources  = ["daemonsets", "deployments", "deployments/rollback", "deployments/scale", "ingresses", "networkpolicies", "replicasets", "replicasets/scale", "replicationcontrollers/scale"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["policy"]
          resources  = ["poddisruptionbudgets"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = ["networking.k8s.io"]
          resources  = ["ingresses", "networkpolicies"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = [""]
          resources  = ["events"]
          verbs      = ["create", "patch", "update"]
        },
        {
          api_groups = [""]
          resources  = ["configmaps", "endpoints", "persistentvolumeclaims", "persistentvolumeclaims/status", "pods", "replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts", "services", "services/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = [""]
          resources  = ["bindings", "events", "limitranges", "namespaces/status", "pods/log", "pods/status", "replicationcontrollers/status", "resourcequotas", "resourcequotas/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = [""]
          resources  = ["namespaces"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["apps"]
          resources  = ["controllerrevisions", "daemonsets", "daemonsets/status", "deployments", "deployments/scale", "deployments/status", "replicasets", "replicasets/scale", "replicasets/status", "statefulsets", "statefulsets/scale", "statefulsets/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["autoscaling"]
          resources  = ["horizontalpodautoscalers", "horizontalpodautoscalers/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["batch"]
          resources  = ["cronjobs", "cronjobs/status", "jobs", "jobs/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["extensions"]
          resources  = ["daemonsets", "daemonsets/status", "deployments", "deployments/scale", "deployments/status", "ingresses", "ingresses/status", "networkpolicies", "replicasets", "replicasets/scale", "replicasets/status", "replicationcontrollers/scale", "replicationcontrollers/scale"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["policy"]
          resources  = ["poddisruptionbudgets", "poddisruptionbudgets/status"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["networking.k8s.io"]
          resources  = ["ingresses", "ingresses/status", "networkpolicies"]
          verbs      = ["get", "list", "watch"]
        },
        {
          api_groups = ["rbac.authorization.k8s.io"]
          resources  = ["rolebindings", "roles"]
          verbs      = ["create", "get", "list", "patch", "update", "watch"]
        },
      ]
    },
  ]

  kms = {
    enabled  = true
    location = "asia-northeast1"
    keyring  = "prod-tokyo"
  }

  import_map_deployer_bucket = {
    "student-coach-e1e95" : {
      project_id    = "student-coach-e1e95"
      bucket_name   = "import-map-deployer-production"
      location      = "asia"
      storage_class = "STANDARD"

      cors = {
        origins = ["*"]
        max_age_seconds = 3600
        response_header = ["*"]
        methods         = ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"]
      }
    }
  }
}
