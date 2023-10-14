locals {
  build_bot_iam_roles = flatten([
    for project, roles in try(var.build_bot.iam.roles, {}) : [
      for role in roles : {
        grant_project = project
        role          = role
      }
    ]
  ])
}

# Create build bot service account
resource "google_service_account" "build_bot_service_account" {
  count = try(var.build_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.build_bot.iam.name)
  display_name = "Build Bot"
  description  = <<EOF
Service account for building and pushing Docker images to registry.
Managed by Terraform.
EOF
  project      = var.build_bot.iam.project_id
}

# Grant build bot service account the specified IAM roles
resource "google_project_iam_member" "build_bot" {
  for_each = {
    for pr in try(local.build_bot_iam_roles, {}) : "${pr.grant_project}/${pr.role}" => pr
  }
  project = each.value.grant_project
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.build_bot_service_account[0].email}"
}

# Provider for the build bot
resource "google_iam_workload_identity_pool_provider" "build_bot" {
  count = try(var.build_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.build_bot.wif.provider_id
  display_name                       = var.build_bot.wif.provider_display_name
  description                        = var.build_bot.wif.provider_description
  attribute_condition                = var.build_bot.wif.attribute_condition
  attribute_mapping                  = var.build_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.build_bot.wif.allowed_audiences
    issuer_uri        = var.build_bot.wif.issuer_uri
  }
}

# Grant build bot to use the provider based on github_repository
resource "google_service_account_iam_member" "build_bot_wif" {
  count = length(google_service_account.build_bot_service_account) == 1 && length(google_iam_workload_identity_pool_provider.build_bot) == 1 ? 1 : 0

  service_account_id = google_service_account.build_bot_service_account[0].id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.build_bot.wif.github_repository}"
}

# This acts as a data source to retrieve the address to the bucket that backs the Container Registry.
#
# See https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/container_registry
# > Ensures that the Google Cloud Storage bucket that backs Google Container Registry exists.
# > Creating this resource will create the backing bucket if it does not exist, or do nothing if the bucket already exists. 
# > Destroying this resource does NOT destroy the backing bucket.
#
# It is created only when build bot SA is created, thus learner project only.
resource "google_container_registry" "asia_gcr_io" {
  count = length(google_service_account.build_bot_service_account) == 1 || length(google_service_account.integration_test_bot_service_account) == 1 ? 1 : 0

  project  = var.project_id
  location = "ASIA"
}

# Allow build bot to access to Container Registry's bucket (e.g. `docker pull`)
resource "google_storage_bucket_iam_member" "build_bot" {
  count = length(google_service_account.build_bot_service_account)

  bucket = google_container_registry.asia_gcr_io[0].id
  role   = "roles/storage.legacyBucketWriter"
  member = "serviceAccount:${google_service_account.build_bot_service_account[0].email}"
}
