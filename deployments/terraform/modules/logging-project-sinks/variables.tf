variable "project_id" {
  type = string
}

variable "sinks" {
  type = list(object({
    name        = string
    destination = string
    description = optional(string)
    filter      = optional(string)
    exclusions = list(object({
      name        = string
      filter      = string
      description = optional(string)
    }))

    disabled               = optional(bool, false)
    unique_writer_identity = optional(bool, false)
  }))

  default     = null
  description = "Create project sink"
}
