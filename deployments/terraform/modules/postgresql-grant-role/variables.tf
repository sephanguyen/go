variable "project_id" {
  type = string
}

variable "env" {
  type = string
}

variable "postgresql_instance" {
  type        = string
  description = "The Postgresql instance used to connect to and grant permissions."
}

variable "postgresql_instance_port" {
  type        = map(string)
  description = "The mappinng port of Postgresql instances that the sql proxy set to listen on."
}

variable "access_level_to_grant_by_pass_rls" {
  type        = list(string)
  description = "The access level which can be granted bypass_rls_role"
}

variable "bypass_rls_role_write_privileges_enabled" {
  type        = bool
  default     = false
  description = "If true, bypass_rls_role has write access. Otherwise, it has read-only access."
}

variable "postgresql_read_only_role_name" {
  type        = string
  description = "Postgresql role used to grant members that can only read from databases."
}

variable "postgresql_read_write_role_name" {
  type        = string
  description = "Postgresql role used to grant members that can read from and write to databases."
}

variable "role_by_access_level" {
  # type = map(map(map(list(string))))
  type = map(map(map(object({
    can_read_databases  = optional(bool)
    can_write_databases = optional(bool)
    custom_roles        = list(string)
  }))))
}

variable "member_by_access_level" {
  type = map(map(list(string)))
}

variable "postgresql_use_predefined_roles" {
  type        = bool
  default     = false
  description = <<EOF
If true, grant members using PostgreSQL's predefined roles (https://www.postgresql.org/docs/current/predefined-roles.html)
instead of using custom roles. Requires PostgreSQL version 14 or above.
EOF
}
