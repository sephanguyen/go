locals {
  mfe_deploy_bot_iam_roles = flatten([
    for project, roles in try(var.mfe_deploy_bot.iam.roles, {}) : [
      for role in roles : {
        grant_project = project
        role          = role
      }
    ]
  ])
}

# Provider for the mfe deploy bot
resource "google_iam_workload_identity_pool_provider" "mfe_deploy_bot" {
  count = try(var.mfe_deploy_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.mfe_deploy_bot.wif.provider_id
  display_name                       = var.mfe_deploy_bot.wif.provider_display_name
  description                        = var.mfe_deploy_bot.wif.provider_description
  attribute_condition                = var.mfe_deploy_bot.wif.attribute_condition
  attribute_mapping                  = var.mfe_deploy_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.mfe_deploy_bot.wif.allowed_audiences
    issuer_uri        = var.mfe_deploy_bot.wif.issuer_uri
  }
}


# Create integration test bot service account
resource "google_service_account" "mfe_deploy_bot_service_account" {
  count = try(var.mfe_deploy_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.mfe_deploy_bot.iam.name)
  display_name = "MFE Upload Artifacts"
  description  = "Service account for running upload artifacts for MFE. Managed by Terraform."
  project      = var.mfe_deploy_bot.iam.project_id
}

# Grant mfe deploy bot to use the provider based on github_repository
resource "google_service_account_iam_member" "mfe_deploy_bot_wif" {
  count = length(google_service_account.mfe_deploy_bot_service_account) == 1 && length(google_iam_workload_identity_pool_provider.mfe_deploy_bot) == 1 ? 1 : 0

  service_account_id = google_service_account.mfe_deploy_bot_service_account[0].id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.mfe_deploy_bot.wif.github_repository}"
}


# Grant mfe_deploy_bot service account the specified IAM roles
resource "google_project_iam_member" "mfe_deploy_bot" {
  for_each = {
    for pr in try(local.mfe_deploy_bot_iam_roles, {}) : "${pr.grant_project}/${pr.role}" => pr
  }
  project = each.value.grant_project
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.mfe_deploy_bot_service_account[0].email}"
}

data "google_storage_bucket" "import_map_deployer" {
  count = length(var.mfe_deploy_bot.iam.bucket_names)

  name = var.mfe_deploy_bot.iam.bucket_names[count.index]
}

resource "google_storage_bucket_iam_member" "mfe_deploy_bot" {
  count = length(data.google_storage_bucket.import_map_deployer)

  bucket = data.google_storage_bucket.import_map_deployer[count.index].name
  role   = "roles/storage.legacyBucketWriter"
  member = "serviceAccount:${google_service_account.mfe_deploy_bot_service_account[0].email}"
}
