/**
* This module:
* - creates preproduction namespaces in Kubernetes
* - grant the deploy bot (specified from input) admin role to those namespaces
* - also grants tech-func-platform@manabie.com admin role to those namespaces (so that I can work with them for now)
*
* This module is highly similar to modules/apps/gke_rbac.tf. TODO: merge them
*/

terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.6.1"
    }
  }
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "https://${var.gke_endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(var.gke_ca_cert)
}
