variable "project_id" {
  description = "The project ID for the resource"
  type        = string
}

variable "org" {
  description = "Organization (abbrev.)"
  type        = string
}

variable "gke_endpoint" {
  description = "The GKE endpoint"
  type        = string
}

variable "gke_ca_cert" {
  description = "The GKE CA certificate"
  type        = string
}

variable "dorp_deploy_bot_service_account" {
  description = "Object of deploy bot's service account created from jp-partners project"
  type = object({
    id    = string
    email = string
  })
}

variable "configure_common_namespaces" {
  description = "Whether to configure permission in common namespaces for deploy bot"
  type        = bool
}

variable "configure_dorp_namespaces" {
  description = "Whether to enable creating and configuring preproduction namespaces"
  type        = bool
}
