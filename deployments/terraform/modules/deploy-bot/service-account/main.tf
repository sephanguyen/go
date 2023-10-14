/**
* This module:
* - creates a service account for the deploy bot
* - adds the deploy bot's SA Key to Github Action's secrets
*/
terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 4.0"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = "manabie-com"
}
resource "google_service_account" "dorp_deploy_bot" {
  project     = var.project_id
  account_id  = "dorp-deploy-bot"
  description = "Deploy bot for Github Action in preproduction environment"
}

resource "google_service_account_key" "dorp_deploy_bot" {
  service_account_id = google_service_account.dorp_deploy_bot.name
}

output "dorp_deploy_bot_service_account" {
  value = {
    id    = google_service_account.dorp_deploy_bot.id
    email = google_service_account.dorp_deploy_bot.email
  }
  description = "dorp-deploy-bot service account"
}
