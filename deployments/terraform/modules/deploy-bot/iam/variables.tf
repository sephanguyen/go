variable "project_id" {
  description = "The project ID for the resource"
  type        = string
}

variable "dorp_deploy_bot_service_account" {
  description = "Object of deploy bot's service account created from jp-partners project"
  type = object({
    id    = string
    email = string
  })
}

variable "production_databases" {
  description = "List of production databases"
  type = list(object({
    project_id  = string
    instance_id = string
  }))
}

variable "preproduction_databases" {
  description = "List of preproduction databases (whose data will be cloned from production)"
  type = list(object({
    project_id  = string
    instance_id = string
  }))
}
