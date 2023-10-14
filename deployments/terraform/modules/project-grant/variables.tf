variable "project_id" {
  type = string
}

variable "env" {
  type = string
}

variable "role_by_access_level" {
  type = map(map(map(object({
    can_read_databases  = optional(bool)
    can_write_databases = optional(bool)
    custom_roles        = list(string)
  }))))
}

variable "member_by_access_level" {
  type = map(map(list(string)))
}

variable techleads {
  type = list(string)
}

variable "techlead_roles" {
  type = map(map(list(string)))
}
