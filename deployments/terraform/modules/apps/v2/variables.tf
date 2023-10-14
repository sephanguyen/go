variable "project_id" {
  type = string
}

variable "service_accounts" {
  type = list(object({
    name                = string
    project             = string
    roles               = map(list(string))
    identity_namespaces = list(string)
    bucket_roles        = optional(map(list(string)))

    impersonations = optional(list(object({
      name    = string
      project = string
      role    = string
    })))
  }))
}

variable "kms_keys" {
  type = map(object({
    key_ring        = string
    rotation_period = string
    owner           = string
    encrypter       = string
    decrypters = list(object({
      service_account_project = string
      service_account_name    = string
    }))
  }))

  default = {}
}

variable "gke_endpoint" {
  type    = string
  default = ""
}

variable "gke_ca_cert" {
  type    = string
  default = ""
}

variable "gke_identity_namespace" {
  type = string
}

variable "gke_rbac" {
  type = object({
    enabled = bool
    policies = list(object({
      kind       = string
      group      = string
      role_kind  = string
      role_name  = string
      namespaces = list(string)
    }))
  })

  default = {
    enabled  = false
    policies = []
  }
}

variable "gke_backup_plan" {
  type = list(object({
    project  = string
    cluster  = string
    name     = string
    location = string
    retention_policy = object({
      backup_delete_lock_days = number
      backup_retain_days      = number
    })
    cron_schedule = string
    backup_config = object({
      include_volume_data = bool
      include_secrets     = bool
      selected_applications = list(object({
        namespace = string
        name      = string
      }))
    })
  }))

  default = []
}

variable "rbac_roles" {
  type = object({
    enabled = bool
    policies = map(map(list(object({
      kind       = string
      group      = string
      namespaces = list(string)
      role_kind  = string
      role_name  = string

      rules = optional(list(object({
        api_groups = list(string)
        resources  = list(string)
        verbs      = list(string)
      })))
    }))))
  })

  default = {
    enabled  = false
    policies = {}
  }
}

variable "create_storage_hmac_key" {
  type = bool
}

variable "bucket_name" {
  type        = string
  default     = ""
  description = "Cloud Storage bucket for backend services, usually created by platforms module."
}

variable "cloudconvert" {
  type = object({
    service_account = string
    bucket          = string
  })

  default     = null
  description = "Service account for Cloudconvert service to upload converted images. It will be granted the storage.objects.create permission on the specify bucket."
}
