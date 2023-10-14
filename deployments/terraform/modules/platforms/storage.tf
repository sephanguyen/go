resource "google_storage_bucket" "backend_bucket" {
  count = var.backend_bucket.enabled ? 1 : 0

  project       = var.project_id
  name          = var.backend_bucket.bucket_name
  location      = var.backend_bucket.location
  storage_class = var.backend_bucket.storage_class

  uniform_bucket_level_access = var.backend_bucket.uniform_bucket_level_access

  dynamic "cors" {
    for_each = var.backend_bucket.cors != null ? var.backend_bucket.cors : []

    content {
      max_age_seconds = cors.value["max_age_seconds"]
      method          = cors.value["method"]
      origin          = cors.value["origin"]
      response_header = cors.value["response_header"]
    }
  }

  versioning {
    enabled = var.backend_bucket.versioning_enabled
  }
}

resource "google_storage_bucket_iam_member" "public_read_backend_bucket" {
  count = var.backend_bucket.enabled ? 1 : 0

  bucket = google_storage_bucket.backend_bucket[0].name
  role   = "roles/storage.objectViewer"
  member = "allUsers"

  depends_on = [google_storage_bucket.backend_bucket]
}


# bucket for import-map-deployer
# Define a variable to store the allowed origins
# Create a Google Cloud Storage bucket
resource "google_storage_bucket" "import_map_deployer_bucket" {
  for_each = var.import_map_deployer_bucket

  project       = each.value.project_id
  name          = each.value.bucket_name
  location      = each.value.location
  storage_class = each.value.storage_class

  # Allow public access to the bucket
  uniform_bucket_level_access = true

  cors {
    # if each.value.cors.origins == ["*"] dont need to concat localhost
    origin           = each.value.cors.origins[0] == "*" ? ["*"] : concat(each.value.cors.origins, ["http://localhost:*", "https://localhost:*", "https://backoffice.local.manabie.io:31600"])
    method          = each.value.cors.methods
    response_header = each.value.cors.response_header
    max_age_seconds = each.value.cors.max_age_seconds
  }

  dynamic "lifecycle_rule" {
    for_each = { for v in each.value.lifecycle_rule : "${v.action.type}_${v.condition.age}" => v }
    content {
      condition {
        age            = lifecycle_rule.value.condition.age
        matches_prefix = lifecycle_rule.value.condition.matches_prefix
      }
      action {
        type = lifecycle_rule.value.action.type
      }
    }
  }
}

# Grant storage.objects.get permission to allUsers
resource "google_storage_bucket_iam_member" "read_all_import_map_deployer_bucket" {
  for_each = var.import_map_deployer_bucket

  bucket = each.value.bucket_name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}