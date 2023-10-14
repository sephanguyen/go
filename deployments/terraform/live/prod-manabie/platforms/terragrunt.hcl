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

  backend_bucket = {
    enabled       = true
    bucket_name   = "manabie-vn-backend"
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

    versioning_enabled          = false
    uniform_bucket_level_access = false
  }

  gke = {
    enabled        = true
    cluster_name   = "manabie"
    region         = "asia-southeast1"
    regional       = true
    zones          = ["asia-southeast1-a", "asia-southeast1-b", "asia-southeast1-c"]
    security_group = "gke-security-groups@manabie.com"

    kubernetes_version = "1.20.10-gke.1600"
    release_channel    = "STABLE"

    network_name      = dependency.vpc.outputs.network_name
    subnetwork_name   = dependency.vpc.outputs.network_name
    ip_range_pods     = "gke-range-pods"
    ip_range_services = "gke-range-services"

    create_service_account = false
    service_account        = "tf-gke-manabie-675f@student-coach-e1e95.iam.gserviceaccount.com"

    gce_pd_csi_driver = true

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
        name         = "n2d-standard-2-c"
        machine_type = "n2d-standard-2"
        autoscaling  = false
        node_count   = 1
        min_count    = null
        max_count    = null
        image_type   = "COS_CONTAINERD"
        spot         = false
      },
    ]

    node_pools_labels = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-2 = {
        cluster_name = "manabie"
        node_pool    = "n2d-standard-2"
      }
    }

    node_pools_metadata = {
      default_values = {
        cluster_name = false
        node_pool    = false
      }

      n2d-standard-4 = {
        cluster_name = "manabie"
        node_pool    = "n2d-standard-2"
      }
    }

    node_pools_taints          = null
    node_pools_resource_labels = null

    node_pools_tags = {
      default_values = [false, false]
    }

    maintenance_start_time = "1970-01-01T17:00:00Z"
    maintenance_end_time   = "1970-01-01T21:00:00Z"
    maintenance_recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"
  }

  gke_enable_platforms_monitoring = true

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
        role_name  = "edit"
        namespaces = []
      },
      {
        kind       = "User"
        group      = "huubang.nguyen@manabie.com"
        role_kind  = "ClusterRole"
        role_name  = "cluster-admin"
        namespaces = []
      },
    ]
  }

  kms = {
    enabled  = true
    location = "asia-southeast1"
    keyring  = "manabie"
  }
}
