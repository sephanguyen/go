# This is referenced from terraform-google-modules/github-actions-runners/google//modules/gh-oidc
# But since we want to share the pool for all Github-related workflows, we can't use that module
# because that module always creates a new pool for each module usage.

# Create a pool for all github's related workflows.
resource "google_iam_workload_identity_pool" "github" {
  provider                  = google-beta
  project                   = var.project_id
  workload_identity_pool_id = var.pool_id
  display_name              = "Github Action Pool"
  description               = "Used by Github Action workflows. Managed by Terraform."
  disabled                  = false
}

# Provider for the deploy bot
resource "google_iam_workload_identity_pool_provider" "docker_bot" {
  count = var.deploy_bot != null ? 1 : 0 # uat-manabie project doesn't have deploy bot

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.deploy_bot.wif.provider_id
  display_name                       = var.deploy_bot.wif.provider_display_name
  description                        = var.deploy_bot.wif.provider_description
  attribute_condition                = var.deploy_bot.wif.attribute_condition
  attribute_mapping                  = var.deploy_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.deploy_bot.wif.allowed_audiences
    issuer_uri        = var.deploy_bot.wif.issuer_uri
  }
}

# Grant deploy bot to use the provider based on github_repository
resource "google_service_account_iam_member" "deploy_bot_wif" {
  count = var.deploy_bot != null ? 1 : 0

  service_account_id = var.deploy_bot.service_account_id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.deploy_bot.wif.github_repository}"
}

resource "google_service_account" "deploy_bot_service_account" {
  count = try(var.deploy_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.deploy_bot.iam.name)
  display_name = "Deploy Bot for Github Actions"
  description  = "Deploy Bot for Github Actions. Managed by Terraform."
  project      = var.deploy_bot.iam.project_id
}

# Grant deploy bot service account the specified IAM roles
resource "google_project_iam_member" "deploy_bot" {
  for_each = try(toset(var.deploy_bot.iam.roles), {})

  project = var.deploy_bot.iam.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.deploy_bot_service_account[0].email}"
}
