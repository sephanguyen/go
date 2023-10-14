variable "project_id" {
  description = "The project ID for the resource"
  type        = string
}

variable "slack_channel" {
  description = "slack channel for alerts"
  type        = string
}

variable "slack_auth_token" {
  description = "slack auth token"
  type        = string
}
