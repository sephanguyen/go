locals {
  kms_encrypters = flatten([
    for key, val in var.kms_keys : [
      for enc in val.encrypters : {
        key = key
        enc = enc
      }
    ]
  ])

  kms_decrypters = flatten([
    for key, val in var.kms_keys : [
      for dec in val.decrypters : {
        key = key
        dec = dec
      }
    ]
  ])
}

resource "google_kms_key_ring" "key_ring" {
  project  = var.project_id
  name     = var.key_ring.name
  location = var.key_ring.location
}

resource "google_kms_crypto_key" "keys" {
  for_each = toset(keys(var.kms_keys))

  name            = each.value
  key_ring        = google_kms_key_ring.key_ring.id
  rotation_period = var.kms_keys[each.value].rotation_period

  lifecycle {
    prevent_destroy = false
  }
}
data "google_service_account" "kms_encrypters" {
  for_each   = toset([for k in local.kms_encrypters : k.enc])
  account_id = each.value
}
data "google_service_account" "kms_decrypters" {
  for_each   = toset([for k in local.kms_decrypters : k.dec])
  account_id = each.value
}
resource "google_kms_crypto_key_iam_member" "kms_owners" {
  for_each = toset(keys(var.kms_keys))

  role          = "roles/owner"
  crypto_key_id = google_kms_crypto_key.keys[each.value].id
  member        = var.kms_keys[each.value].owner
}

resource "google_cloud_identity_group" "techlead_groups" {
  for_each = var.create_google_groups ? toset(keys(var.kms_keys)) : []

  parent = "customers/C00ziiz00"
  group_key {
    id = var.kms_keys[each.value].encrypter_decrypter
  }
  labels = {
    "cloudidentity.googleapis.com/groups.discussion_forum" = ""
  }
  description = "Group ${var.kms_keys[each.value].encrypter_decrypter} managed by Terraform"
}

resource "google_kms_crypto_key_iam_member" "kms_encrypter_decrypters" {
  for_each = toset(keys(var.kms_keys))

  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  crypto_key_id = google_kms_crypto_key.keys[each.value].id
  member        = "group:${var.kms_keys[each.value].encrypter_decrypter}"
  depends_on    = [google_cloud_identity_group.techlead_groups]
}

resource "google_kms_crypto_key_iam_member" "kms_encrypters" {
  for_each = {
    for ke in local.kms_encrypters :
    "${ke.key}.${ke.enc}" => ke
  }

  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.keys[each.value.key].id
  member        = "serviceAccount:${each.value.enc}"
}

resource "google_kms_crypto_key_iam_member" "kms_decrypters" {
  for_each = {
    for kd in local.kms_decrypters :
    "${kd.key}.${kd.dec}" => kd
  }

  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.keys[each.value.key].id
  member        = "serviceAccount:${each.value.dec}"
}
