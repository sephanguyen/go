terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.10.0"
    }
  }
}

data "google_client_config" "default" {}

provider "kubernetes" {
  # config_path = "~/.kube/config"
  host                   = var.gke.enabled ? "https://${module.gke[0].endpoint}" : ""
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = var.gke.enabled ? base64decode(module.gke[0].ca_certificate) : ""
}
