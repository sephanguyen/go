variable "project_id" {
  type = string
}

variable "roles" {
  type = list(object({
    id          = string
    title       = string
    description = string

    base_roles  = list(string)
    permissions = list(string)
  }))
}
