output "storage_hmac_key" {
  value = length(google_storage_hmac_key.storage) > 0 ? {
    access_id  = google_storage_hmac_key.storage[0].access_id
    secret_key = google_storage_hmac_key.storage[0].secret
  } : null
  sensitive = true
}

output "cloudconvert_service_account_key" {
  value     = var.cloudconvert != null ? google_service_account_key.cloudconvert[0].private_key : null
  sensitive = true
}
