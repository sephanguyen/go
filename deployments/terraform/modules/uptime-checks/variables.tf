variable "project_id" {
  description = "The project ID to put the uptime check in"
  type        = string
}

variable "hasura_host" {
  type        = string
  description = "Hasura root host"
}

variable "hasura_port" {
  type        = string
  default     = "443"
  description = "Hasura root port"
}

variable "hasura_paths" {
  type        = list(string)
  description = "Hasura service paths to do health check"
}

variable "https_check" {
  description = "HTTPS for inbound calls"
  type = map(object({
    port           = string
    host           = string
    request_method = optional(string, "GET")
    name           = string
    paths          = list(string)
    content_type   = optional(string, "TYPE_UNSPECIFIED")
    body           = optional(string, "")

    content_matchers = optional(object({
      content = optional(string)
      matcher = optional(string)
      }), {
      content = "OK"
      matcher = "CONTAINS_STRING"
    })

    accepted_response_status_codes = optional(object({
      status_value = optional(number)
      status_class = optional(string)
      }), {
      status_value = 200
      status_class = "STATUS_CLASS_2XX"
    })
  }))
}
