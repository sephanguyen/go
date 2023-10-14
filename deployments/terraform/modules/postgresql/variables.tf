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

Because some objects are created by the users, not postgres (e.g. hasura, migration, etc...),
`GRANT <role> to <atlantis service account>` will also be run for hasura and migration users.
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

variable "postgresql_bypass_rls_roles" {
  type    = list(string)
  default = []
}

variable "postgresql_replication_roles" {
  type    = list(string)
  default = []
}

variable "pgaudit_enabled" {
  type        = bool
  default     = false
  description = "Whether to enable PgAudit extension for PostgreSQL."
}

variable "postgresql_statement_timeout" {
  type = list(object({
    user              = string
    statement_timeout = string
  }))

  default = []
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
