locals {
  unleash_decryptor_bot_iam_roles = flatten([
    for project, roles in try(var.unleash_decryptor_bot.iam.roles, {}) : [
      for role in roles : {
        grant_project = project
        role          = role
      }
    ]
  ])
}

# Create unleash decryptor bot service account
resource "google_service_account" "unleash_decryptor_bot_service_account" {
  count = try(var.unleash_decryptor_bot.iam, null) != null ? 1 : 0

  account_id   = lower(var.unleash_decryptor_bot.iam.name)
  display_name = "Unleash Decryptor Bot"
  description  = "Service account to decrypt Unleash's admin token secret file. Managed by Terraform."
  project      = var.unleash_decryptor_bot.iam.project_id
}

# Grant unleash decryptor bot service account the specified IAM roles
resource "google_project_iam_member" "unleash_decryptor_bot" {
  for_each = {
    for pr in try(local.unleash_decryptor_bot_iam_roles, {}) : "${pr.grant_project}/${pr.role}" => pr
  }
  project = each.value.grant_project
  role    = each.value.role
  member  = "serviceAccount:${google_service_account.unleash_decryptor_bot_service_account[0].email}"
}

# Provider for the unleash decryptor bot
resource "google_iam_workload_identity_pool_provider" "unleash_decryptor_bot" {
  count = try(var.unleash_decryptor_bot.wif, null) != null ? 1 : 0

  provider                           = google-beta
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = var.unleash_decryptor_bot.wif.provider_id
  display_name                       = var.unleash_decryptor_bot.wif.provider_display_name
  description                        = var.unleash_decryptor_bot.wif.provider_description
  attribute_condition                = var.unleash_decryptor_bot.wif.attribute_condition
  attribute_mapping                  = var.unleash_decryptor_bot.wif.attribute_mapping
  oidc {
    allowed_audiences = var.unleash_decryptor_bot.wif.allowed_audiences
    issuer_uri        = var.unleash_decryptor_bot.wif.issuer_uri
  }
}

# Grant unleash decryptor bot to use the provider based on github_repository
resource "google_service_account_iam_member" "unleash_decryptor_bot_wif" {
  count = length(google_service_account.unleash_decryptor_bot_service_account) == 1 && length(google_iam_workload_identity_pool_provider.unleash_decryptor_bot) == 1 ? 1 : 0

  service_account_id = google_service_account.unleash_decryptor_bot_service_account[0].id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${var.unleash_decryptor_bot.wif.github_repository}"
}

# This is the key that is used to encrypt the admin token file
data "google_kms_key_ring" "manabie" {
  count = length(google_service_account.unleash_decryptor_bot_service_account)

  project  = var.project_id
  name     = "manabie"
  location = "asia-southeast1"
}

data "google_kms_crypto_key" "prod_manabie" {
  count = length(google_service_account.unleash_decryptor_bot_service_account)

  name     = "prod-manabie"
  key_ring = data.google_kms_key_ring.manabie[0].id
}

# Grant unleash decryptor bot permissions to decrypt admin token file
resource "google_kms_crypto_key_iam_member" "unleash" {
  count = length(google_service_account.unleash_decryptor_bot_service_account)

  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = data.google_kms_crypto_key.prod_manabie[0].id
  member        = "serviceAccount:${google_service_account.unleash_decryptor_bot_service_account[0].email}"
}
