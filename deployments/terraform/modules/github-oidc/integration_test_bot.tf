locals {
  integration_test_bot_iam_roles = flatten([
    for project, roles in try(var.integration_test_bot.iam.roles, {}) : [
      for role in roles : {
        grant_project = project
        role          = role
      }
    ]
  ])
}

# Create integration test bot service account
resource "google_service_account" "integration_test_bot_service_account" {
  count = try(var.integration_test_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.integration_test_bot.iam.name)
  display_name = "Integration Test Bot"
  description  = "Service account for running backend's integration test. Managed by Terraform."
  project      = var.integration_test_bot.iam.project_id
}

# Grant integration test bot service account the specified IAM roles
resource "google_project_iam_member" "integration_test_bot" {
  for_each = {
    for pr in try(local.integration_test_bot_iam_roles, {}) : "${pr.grant_project}/${pr.role}" => pr
  }
  project = each.value.grant_project
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.integration_test_bot_service_account[0].email}"
}

# Provider for the integration test bot
resource "google_iam_workload_identity_pool_provider" "integration_test_bot" {
  count = try(var.integration_test_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.integration_test_bot.wif.provider_id
  display_name                       = var.integration_test_bot.wif.provider_display_name
  description                        = var.integration_test_bot.wif.provider_description
  attribute_condition                = var.integration_test_bot.wif.attribute_condition
  attribute_mapping                  = var.integration_test_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.integration_test_bot.wif.allowed_audiences
    issuer_uri        = var.integration_test_bot.wif.issuer_uri
  }
}

# Grant integration test bot to use the provider based on github_repository
resource "google_service_account_iam_member" "integration_test_bot_wif" {
  count = length(google_service_account.integration_test_bot_service_account) == 1 && length(google_iam_workload_identity_pool_provider.integration_test_bot) == 1 ? 1 : 0

  service_account_id = google_service_account.integration_test_bot_service_account[0].id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.integration_test_bot.wif.github_repository}"
}

# Allow integration test bot to access to Container Registry's bucket (e.g. `docker pull`)
resource "google_storage_bucket_iam_member" "integration_test_bot" {
  count = length(google_service_account.integration_test_bot_service_account)

  bucket = google_container_registry.asia_gcr_io[0].id
  role   = "roles/storage.legacyBucketReader"
  member = "serviceAccount:${google_service_account.integration_test_bot_service_account[0].email}"
}

// Repository that contains docker images used on CI.
resource "google_artifact_registry_repository" "ci" {
  count = length(google_service_account.integration_test_bot_service_account)

  project       = "student-coach-e1e95" # might conflict if the count above fails to prevent multiple creation
  repository_id = "ci"
  format        = "DOCKER"
  location      = "asia-southeast1"
  description   = "CI registry managed by Terraform."
}
