# Provider for the mfe deploy bot
resource "google_iam_workload_identity_pool_provider" "get_release_tag_bot" {
  count = try(var.get_release_tag_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.get_release_tag_bot.wif.provider_id
  display_name                       = var.get_release_tag_bot.wif.provider_display_name
  description                        = var.get_release_tag_bot.wif.provider_description
  attribute_condition                = var.get_release_tag_bot.wif.attribute_condition
  attribute_mapping                  = var.get_release_tag_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.get_release_tag_bot.wif.allowed_audiences
    issuer_uri        = var.get_release_tag_bot.wif.issuer_uri
  }
}


# Create integration test bot service account
resource "google_service_account" "get_release_tag_bot_service_account" {
  count = try(var.get_release_tag_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.get_release_tag_bot.iam.name)
  display_name = "Get release tag"
  description  = "Service account for running upload artifacts for get release tag. Managed by Terraform."
  project      = var.get_release_tag_bot.iam.project_id
}

# Grant get release tag deploy bot to use the provider
resource "google_service_account_iam_member" "get_release_tag_bot_wif" {
  count = length(google_service_account.get_release_tag_bot_service_account) == 1 && length(google_iam_workload_identity_pool_provider.get_release_tag_bot) == 1 ? 1 : 0

  service_account_id = google_service_account.get_release_tag_bot_service_account[0].id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/*" #All identities in the pool
}

# Grant get_release_tag_bot service account the specified IAM roles
resource "google_project_iam_member" "get_release_tag_bot" {
  for_each = try(toset(var.get_release_tag_bot.iam.roles), {})

  project = var.get_release_tag_bot.iam.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.get_release_tag_bot_service_account[0].email}"
}
