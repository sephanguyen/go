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

variable "postgresql_read_only_role_name" {
  type        = string
  description = "Postgresql role used to grant members that can only read from databases."
}

variable "postgresql_read_write_role_name" {
  type        = string
  description = "Postgresql role used to grant members that can read from and write to databases."
}

variable "postgresql_databases" {
  type = list(string)
}
