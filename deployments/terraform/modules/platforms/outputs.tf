output "postgresql_project" {
  value = var.project_id
}

# In order to keep backward compatibility, we need to keep the old
# postgresql_project output, since there are places still using it.
# We can remove this output later, when we refactor all of those places.
output "postgresql_instance" {
  value = length(keys(var.postgresql)) > 0 ? module.postgresql[keys(var.postgresql)[0]].instance_name : null
}

output "postgresql_instances" {
  value = {
    for name, config in module.postgresql :
    name => config.instance_name
  }
}

output "postgresql_password" {
  value = {
    for name, config in module.postgresql :
    name => config.generated_user_password
  }
  sensitive = true
}

output "gke_cluster_id" {
  value = var.gke.enabled ? module.gke[0].cluster_id : null
}

output "gke_endpoint" {
  value     = var.gke.enabled ? module.gke[0].endpoint : null
  sensitive = true
}

output "gke_cluster_name" {
  value = var.gke.enabled ? module.gke[0].name : null
}

output "gke_ca_cert" {
  value     = var.gke.enabled ? module.gke[0].ca_certificate : null
  sensitive = true
}

output "gke_identity_namespace" {
  value = var.gke.enabled ? module.gke[0].identity_namespace : null
}

output "kms_key_ring" {
  value = var.kms.enabled ? module.kms[0].keyring : null
}

output "bigquery_dataset" {
  value       = var.bigquery.enabled ? module.bigquery[0].bigquery_dataset : null
  description = "Bigquery dataset resource."
}

output "backend_bucket" {
  value = var.backend_bucket.enabled ? google_storage_bucket.backend_bucket[0].name : null
}
