variable "project_id" {
  type = string
}

variable "env" {
  type = string
}

variable "postgresql_project_id" {
  type        = string
  description = "The GCP project that the Postgresql instance belongs to"
}

variable "postgresql_instance" {
  type        = string
  description = "The Postgresql instance used to connect to and grant permissions."
}

variable "postgresql_instance_port" {
  type        = map(string)
  description = "The mappinng port of Postgresql instances that the sql proxy set to listen on."
}

variable "role_by_access_level" {
  type = map(map(map(object({
    can_read_databases  = optional(bool, false)
    can_write_databases = optional(bool, false)
    custom_roles        = list(string)
  }))))
}

variable "member_by_access_level" {
  type = map(map(list(string)))
}
