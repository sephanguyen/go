variable "project_id" {
  type = string
}
variable "exclusions" {
  type = list(object({
    name        = string
    description = optional(string)
    filter      = string
  }))

  default = []
  description = "Create logging project exclusions"
}
