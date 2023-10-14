variable "project_id" {
  type = string
}

variable "postgresql" {
  type = map(object({
    enabled          = bool
    name             = string
    random_suffix    = bool
    region           = string
    zone             = string
    tier             = string
    disk_size        = string
    disk_autoresize  = bool
    database_version = string

    database_flags = list(object({
      name  = string
      value = string
    }))

    insights_config = object({
      query_string_length     = number
      record_application_tags = bool
      record_client_address   = bool
    })

    deletion_protection = bool

    maintenance_window_day          = number
    maintenance_window_hour         = number
    maintenance_window_update_track = string

    backup_location                = string
    backup_start_time              = string
    point_in_time_recovery_enabled = bool
    transaction_log_retention_days = string
    retained_backups               = number
    retention_unit                 = string

    private_network = string
    authorized_networks = optional(list(object({
      name  = string
      value = string
    })), [])
  }))

  default = {}
}

variable "postgresql_alerts" {
  type = object({
    memory_utilization = object({
      duration        = optional(string, "180s")
      comparison      = optional(string, "COMPARISON_GT")
      threshold_value = optional(number, 0.80)
    })
  })

  default = {
    memory_utilization = {
      duration        = "180s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.80
    }
  }

  description = <<EOF
Configuration for postgres' related alerts.

Not all alerts are configurable here, since I only need
to adjust some alerts. Add more if you need to.
EOF
}

variable "kms" {
  type = object({
    enabled  = bool
    location = string
    keyring  = string
  })

  default = {
    enabled  = false
    location = null
    keyring  = null
  }
}

variable "backend_bucket" {
  type = object({
    enabled       = bool
    bucket_name   = string
    location      = string
    storage_class = string

    cors = list(object({
      max_age_seconds = number
      method          = list(string)
      origin          = list(string)
      response_header = list(string)
    }))

    versioning_enabled          = bool
    uniform_bucket_level_access = bool
  })

  default = {
    enabled       = false
    bucket_name   = null
    location      = null
    storage_class = null

    cors = null

    versioning_enabled          = null
    uniform_bucket_level_access = null
  }
}

variable "gke" {
  type = object({
    enabled        = bool
    cluster_name   = string
    region         = string
    regional       = string
    zones          = list(string)
    security_group = string

    kubernetes_version = string
    release_channel    = string

    network_name      = string
    subnetwork_name   = string
    ip_range_pods     = string
    ip_range_services = string

    service_account        = string
    create_service_account = bool

    gce_pd_csi_driver   = bool
    backup_agent_config = optional(bool, false)

    cluster_autoscaling = object({
      enabled             = optional(bool, false)
      autoscaling_profile = optional(string, "BALANCED")
      min_cpu_cores       = optional(number, 0)
      max_cpu_cores       = optional(number, 0)
      min_memory_gb       = optional(number, 0)
      max_memory_gb       = optional(number, 0)
      gpu_resources       = optional(list(object({ resource_type = string, minimum = number, maximum = number })), [])
      auto_repair         = optional(bool, true)
      auto_upgrade        = optional(bool, true)
    })

    node_pools = list(object({
      name           = string
      machine_type   = string
      autoscaling    = bool
      node_count     = number
      min_count      = number
      max_count      = number
      image_type     = string
      spot           = bool
      disk_size_gb   = optional(number, 100)
      disk_type      = optional(string, "pd-standard")
      node_locations = optional(string, "")
    }))

    node_pools_labels          = map(map(string))
    node_pools_metadata        = map(map(string))
    node_pools_tags            = map(list(string))
    node_pools_taints          = map(list(object({ key = string, value = string, effect = string })))
    node_pools_resource_labels = map(map(string))

    network_policy = optional(bool, false)

    maintenance_start_time = string
    maintenance_end_time   = string
    maintenance_recurrence = string
  })

  default = {
    enabled        = false
    cluster_name   = null
    region         = null
    regional       = null
    zones          = null
    security_group = null

    kubernetes_version = null
    release_channel    = null

    network_name      = null
    subnetwork_name   = null
    ip_range_pods     = null
    ip_range_services = null

    service_account        = null
    create_service_account = null

    gce_pd_csi_driver   = null
    backup_agent_config = null

    cluster_autoscaling = null

    node_pools = null

    node_pools_labels          = null
    node_pools_metadata        = null
    node_pools_tags            = null
    node_pools_taints          = null
    node_pools_resource_labels = null

    maintenance_start_time = null
    maintenance_end_time   = null
    maintenance_recurrence = null
  }
}

variable "kubernetes_cluster_roles" {
  type = list(object({
    name = string
    rules = list(object({
      api_groups = list(string)
      resources  = list(string)
      verbs      = list(string)
    }))
  }))

  default = []
}

variable "gke_rbac" {
  type = object({
    enabled = bool
    policies = list(object({
      kind      = string
      group     = string
      role_kind = string
      role_name = string

      # If namespaces are empty, cluster role binding will be created.
      # Otherwise role binding will be created in each namespace.
      namespaces = list(string)
    }))
  })

  default = {
    enabled  = false
    policies = null
  }
}

variable "gke_enable_resources_monitoring" {
  type        = bool
  default     = false
  description = "Enable monitoring for GKE resources, such as CPU and Memory utilization"
}

variable "gke_enable_platforms_monitoring" {
  type        = bool
  default     = false
  description = "Enable monitoring for platforms deployments including Istio, Prometheus, Alertmanager...and so on."
}

variable "bigquery" {
  type = object({
    enabled                    = bool
    delete_contents_on_destroy = bool
    dataset_id                 = string
    dataset_name               = string
    description                = string
    location                   = string
    dataset_labels             = map(string)
  })

  default = {
    enabled                    = false
    dataset_id                 = null
    dataset_labels             = null
    dataset_name               = null
    delete_contents_on_destroy = null
    description                = null
    location                   = null
  }
}

variable "import_map_deployer_bucket" {
  type = map(
    object({
      project_id    = optional(string) // default is key of map
      bucket_name   = optional(string)
      location      = optional(string)
      storage_class = optional(string)

      cors = object({
        origins         = optional(list(string), [])
        max_age_seconds = optional(number, 3600)
        response_header = optional(list(string), ["*"])
        methods         = optional(list(string), ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"])
      })

      lifecycle_rule = optional(list(object({
        condition = object({
          age            = number
          matches_prefix = optional(list(string), ["manabie", "tokyo", "jprep"])
        })
        action = object({
          type = string
        })
      })), [])
    })
  )

  default = {}
}
