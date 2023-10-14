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
  project_id = "student-coach-e1e95"

  postgresql = {
    jp-partners = {
      enabled          = true
      name             = "jp-partners-b04fbb69"
      random_suffix    = false
      region           = "asia-northeast1"
      zone             = "asia-northeast1-c"
      database_version = "POSTGRES_13"
      tier             = "db-custom-16-49152"
      disk_size        = 15
      disk_autoresize  = true

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
        {
          name  = "max_connections"
          value = "1000"
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
      authorized_networks = [
        {
          name  = "prod-tokyo instance outgoing IP address (For LT-43291)"
          value = "35.243.82.178"
        },
        {
          name  = "prod-tokyo-lms-b2dc4508 instance outgoing IP address (For LT-43291)"
          value = "34.146.197.58"
        },
      ]
    }
  }

  backend_bucket = {
    enabled       = true
    bucket_name   = "jp-partners-backend"
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
    cluster_name   = "jp-partners"
    region         = "asia-northeast1"
    regional       = true
    zones          = ["asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.19.14-gke.1900"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = "gke-range-pods"
    ip_range_services = "gke-range-services"

    create_service_account = false
    service_account        = "tf-gke-jp-partners-a73s@student-coach-e1e95.iam.gserviceaccount.com"

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
      },
      {
        name         = "n2d-highmem-4"
        machine_type = "n2d-highmem-4"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 3
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name           = "pool-backend-on-demand"
        machine_type   = "n2d-highmem-2"
        autoscaling    = true
        node_count     = null
        min_count      = 0
        max_count      = 2
        image_type     = "COS_CONTAINERD"
        spot           = false
        node_locations = "asia-northeast1-c"
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
        name         = "n2d-standard-4-spot"
        machine_type = "n2d-standard-4"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },
      {
        name         = "n2d-standard-2-spot"
        machine_type = "n2d-standard-2"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 1
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 100
      },
      {
        name         = "preproduction-n2d-highmem-2-spot"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 0
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

      pool-backend-on-demand = {
        backend-on-demand-node = "true"
      }

      pool-kafka-on-demand = {
        kafka = "true"
      }

      n2d-standard-4 = {
        cluster_name = "jp-partners"
        node_pool    = "n2d-standard-4"
      }

      preproduction-n2d-highmem-2-spot = {
        cluster_name = "jp-partners"
        node_pool    = "preproduction-n2d-highmem-2-spot"
        environment  = "preproduction"
      }
    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-4 = {
        cluster_name = "jp-partners"
        node_pool    = "n2d-standard-4"
      }
    }

    node_pools_taints = {
      all = []

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

      n2d-standard-4-spot = [
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

    node_pools_tags = {
      default_values = [false, false]

      n2d-standard-4 = [
        "gke-jp-partners",
        "gke-jp-partners-n2d-standard-4",
      ]
    }

    node_pools_resource_labels = null

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

  gke_enable_resources_monitoring = true
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
    keyring  = "jp-partners"
  }
}
