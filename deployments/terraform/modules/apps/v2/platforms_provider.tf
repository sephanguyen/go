terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.6.1"
    }

    null = {
      source  = "hashicorp/null"
      version = "3.1.1"
    }
  }
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = var.gke_endpoint != "" ? "https://${var.gke_endpoint}" : ""
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = var.gke_ca_cert != "" ? base64decode(var.gke_ca_cert) : ""
}
