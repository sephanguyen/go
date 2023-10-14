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
  gke = {
    enabled        = true
    cluster_name   = "staging-2"
    region         = "asia-southeast1"
    regional       = false
    zones          = ["asia-southeast1-b"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.21.11-gke.1900"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = ""
    ip_range_services = ""

    create_service_account = false
    service_account        = "tf-gke-staging-p642@staging-manabie-online.iam.gserviceaccount.com"

    gce_pd_csi_driver   = true
    backup_agent_config = true

    network_policy = true

    cluster_autoscaling = {
      enabled             = false
      autoscaling_profile = "BALANCED"
      min_cpu_cores       = 0
      max_cpu_cores       = 0
      min_memory_gb       = 0
      max_memory_gb       = 0
      gpu_resources       = []
    }

    node_pools = [
      # services pools
      {
        name         = "e2-standard-8-ml-spot"
        machine_type = "e2-standard-8"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 3
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
      },
      {
        name         = "e2-medium-spot"
        machine_type = "e2-medium"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 3
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
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
        name         = "n2d-highmem-2-spot"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 2
        image_type   = "COS_CONTAINERD"
        spot         = true
      },
      {
        name         = "n2d-highmem-2-on-demand"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 3
        max_count    = 5
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "n2d-custom-2-24"
        machine_type = "n2d-custom-2-24576-ext"
        autoscaling  = true
        node_count   = null
        min_count    = 1
        max_count    = 1
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
      {
        name         = "n2d-highmem-2-uat-spot"
        machine_type = "n2d-highmem-2"
        autoscaling  = true
        node_count   = null
        min_count    = 3
        max_count    = 4
        image_type   = "COS_CONTAINERD"
        spot         = true
      },

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
        name         = "custom-8-13-runners-spot"
        machine_type = "n2d-custom-8-16384"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
        disk_type    = "pd-balanced"
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
        name         = "n2d-standard-8-runners-balanced-spot"
        machine_type = "n2d-standard-8"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 100
        disk_type    = "pd-balanced"
      },
      {
        name         = "custom-4-20-runners-spot"
        machine_type = "n2d-custom-4-20480"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 100
        disk_type    = "pd-balanced"
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
      {
        name         = "n2d-standard-8-runners-spot"
        machine_type = "n2d-standard-8"
        autoscaling  = true
        node_count   = null
        min_count    = 0
        max_count    = 30
        image_type   = "COS_CONTAINERD"
        spot         = true
        disk_size_gb = 50
        disk_type    = "pd-balanced"
      },
    ]

    node_pools_labels = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      t2d-standard-2-spot = {
        spot = "true"
      }

      n2d-standard-4-on-demand = {
        on-demand = "true"
      }

      n2d-highmem-2-uat-spot = {
        n2d-highmem-2-uat-spot = "true"
      }

      n2d-standard-4-runners-spot = {
        role = "runner"
      }

      e2-highmem-2-runners-on-demand = {
        role = "runner"
      }

      t2d-standard-1-runners-on-demand = {
        role = "runner"
      }

      t2d-standard-1-runners-spot = {
        role = "runner"
      }


      custom-8-13-runners-spot = {
        role = "runner"
      }

      c2d-highcpu-4-runners-spot = {
        role = "runner"
      }

      n2d-highmem-2-runners-spot = {
        role = "runner"
      }

      n2d-standard-8-runners-balanced-spot = {
        role = "runner"
      }

      custom-4-20-runners-spot = {
        role = "runner"
      }

      n2d-standard-8-runners-spot = {
        role = "runner"
      }

    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }
    }

    node_pools_taints = {
      e2-standard-8-ml-spot = [
        {
          key    = "e2-standard-8-ml-scheduling-timetable-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      e2-medium-spot = [
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

      n2d-highmem-2-spot = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

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

      "custom-8-13-runners-spot" = [
        {
          key    = "custom-8-13-runners-spot"
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

      "n2d-standard-8-runners-balanced-spot" = [
        {
          key    = "n2d-standard-8-runners-balanced-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "custom-4-20-runners-spot" = [
        {
          key    = "custom-4-20-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]

      "n2d-highmem-2-uat-spot" = [
        {
          key    = "cloud.google.com/gke-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
        {
          key    = "n2d-highmem-2-uat-spot"
          value  = true
          effect = "NO_SCHEDULE"
        }
      ]

      "n2d-standard-8-runners-spot" = [
        {
          key    = "n2d-standard-8-runners-spot"
          value  = true
          effect = "NO_SCHEDULE"
        },
      ]
    }

    node_pools_resource_labels = {

      n2d-standard-4-runners-spot = {
        role = "runner"
      }

      e2-highmem-2-runners-on-demand = {
        role = "runner"
      }

      t2d-standard-1-runners-on-demand = {
        role = "runner"
      }

      t2d-standard-1-runners-spot = {
        role = "runner"
      }


      custom-8-13-runners-spot = {
        role = "runner"
      }

      c2d-highcpu-4-runners-spot = {
        role = "runner"
      }

      n2d-highmem-2-runners-spot = {
        role = "runner"
      }

      n2d-standard-8-runners-balanced-spot = {
        role = "runner"
      }

      custom-4-20-runners-spot = {
        role = "runner"
      }

      n2d-standard-8-runners-spot = {
        role = "runner"
      }
    }

    node_pools_tags = {
      default_values = [false, false]
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
        role_name  = "admin"
        namespaces = []
      },
      {
        kind       = "Group"
        group      = "tech-squad-architecture@manabie.com"
        role_kind  = "ClusterRole"
        role_name  = "admin"
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
}
