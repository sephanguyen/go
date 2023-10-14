terraform {
  required_version = "~> 1.4.0"

  required_providers {
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "1.18.0"
    }
  }
}

provider "postgresql" {
  host            = "127.0.0.1"
  port            = var.postgresql_instance_port[var.postgresql_instance]
  username        = "atlantis@student-coach-e1e95.iam"
  sslmode         = "disable"
  max_connections = 4
}
