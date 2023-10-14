resource "google_kms_crypto_key" "keys" {
  for_each = toset(keys(var.kms_keys))

  name            = each.value
  key_ring        = var.kms_keys[each.value].key_ring
  rotation_period = var.kms_keys[each.value].rotation_period

  lifecycle {
    prevent_destroy = false
  }
}

resource "google_kms_crypto_key_iam_member" "kms_owners" {
  for_each = toset(keys(var.kms_keys))

  role          = "roles/owner"
  crypto_key_id = "${google_kms_crypto_key.keys[each.value].key_ring}/cryptoKeys/${each.value}"
  member        = var.kms_keys[each.value].owner
}

resource "google_service_account" "kms_encrypters" {
  for_each = toset([
    for key, val in var.kms_keys : val.encrypter
  ])

  project    = var.project_id
  account_id = each.value
}

resource "google_kms_crypto_key_iam_member" "kms_encrypters" {
  for_each = toset(keys(var.kms_keys))

  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = "${google_kms_crypto_key.keys[each.value].key_ring}/cryptoKeys/${each.value}"
  member        = "serviceAccount:${google_service_account.kms_encrypters[var.kms_keys[each.value].encrypter].email}"
}

locals {
  kms_decrypters = flatten([
    for key, val in var.kms_keys : [
      for dec in val.decrypters : {
        key                     = key
        service_account_project = dec.service_account_project
        service_account_name    = dec.service_account_name
      }
      if dec.service_account_name != null
    ]
  ])
}

resource "google_kms_crypto_key_iam_member" "kms_decrypters" {
  for_each = {
    for kd in local.kms_decrypters :
    "${kd.key}.${kd.service_account_project}.${kd.service_account_name}" => kd
  }

  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = "${google_kms_crypto_key.keys[each.value.key].key_ring}/cryptoKeys/${each.value.key}"
  member = "serviceAccount:${lookup(
    google_service_account.service_accounts,
    "${each.value.service_account_project}.${each.value.service_account_name}",
    {}
  ).email}"
}
