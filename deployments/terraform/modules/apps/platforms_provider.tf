terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.6.1"
    }

    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "1.18.0"
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

provider "postgresql" {
  host            = "127.0.0.1"
  port            = var.postgresql_instance_port[var.postgresql.instance]
  username        = "atlantis@student-coach-e1e95.iam"
  sslmode         = "disable"
  max_connections = 4
}
