terraform {
  required_version = "~> 1.4.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.75.1"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "4.75.1"
    }
  }
}
