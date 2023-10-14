# This is referenced from terraform-google-modules/github-actions-runners/google//modules/gh-oidc
# But since we want to share the pool for all Github-related workflows, we can't use that module
# because that module always creates a new pool for each module usage.

# Provider for the deploy bot
resource "google_iam_workload_identity_pool_provider" "dorp_deploy_bot" {
  count = try(var.dorp_deploy_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.dorp_deploy_bot.wif.provider_id
  display_name                       = var.dorp_deploy_bot.wif.provider_display_name
  description                        = var.dorp_deploy_bot.wif.provider_description
  attribute_condition                = var.dorp_deploy_bot.wif.attribute_condition
  attribute_mapping                  = var.dorp_deploy_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.dorp_deploy_bot.wif.allowed_audiences
    issuer_uri        = var.dorp_deploy_bot.wif.issuer_uri
  }
}

# Grant deploy bot to use the provider based on github_repository
resource "google_service_account_iam_member" "dorp_deploy_bot_wif" {
  count = try(var.dorp_deploy_bot.wif, null) != null ? 1 : 0

  service_account_id = var.dorp_deploy_bot.service_account_id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.dorp_deploy_bot.wif.github_repository}"
}
