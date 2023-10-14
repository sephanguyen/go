terraform {
  required_version = "~> 1.4.0"

  required_providers {
    slack = {
      source  = "manabie-com/slack"
      version = "~> 1.2"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = "manabie-com"
}

provider "slack" {
  token = var.slack_token
}

provider "google-beta" {
  billing_project       = var.project_id
  user_project_override = true
}
