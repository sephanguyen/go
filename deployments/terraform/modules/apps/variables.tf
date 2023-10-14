variable "project_id" {
  type = string
}

variable "postgresql_instance_port" {
  type        = map(string)
  description = "The mappinng port of Postgresql instances that the sql proxy set to listen on."
}

variable "postgresql" {
  type = object({
    project_id = string
    instance   = string
    databases  = list(string)
    users      = list(string)
  })

  description = <<EOF
PostgreSQL configurations.
If a database contains suffix "hasura_metadata" in its name, it is considered a database for hasura metadata
(which will have additional setups).
EOF
}

variable "postgresql_user_permissions" {
  type = list(object({
    database    = string
    user        = string
    owner       = optional(string, "postgres")
    schema      = string
    object_type = string
    privileges  = list(string)
    objects     = optional(list(string))
  }))
}

variable "adhoc" {
  type = object({
    grant_enabled = optional(bool, false)
    dbuser        = optional(string)
  })
  default = {
    grant_enabled = false
  }
  description = <<EOF
Whether to grant migration roles to adhoc db user (so that adhoc account
have permissions to execute DDLs).
EOF
}

variable "postgresql_bypass_rls_roles" {
  type    = list(string)
  default = []
}

variable "postgresql_replication_roles" {
  type    = list(string)
  default = []
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

variable "postgresql_statement_timeout" {
  type = list(object({
    user              = string
    statement_timeout = string
  }))

  default = []
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

variable "pgaudit_enabled" {
  type        = bool
  default     = false
  description = "Whether to enable PgAudit extension for PostgreSQL."
}
