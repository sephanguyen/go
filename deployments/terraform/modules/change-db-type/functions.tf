# Setup the root directory of where the source code will be stored.
locals {
  root_dir = abspath("src")
}

data "google_project" "project" {}

# Zip up our code so that we can store it for deployment.
data "archive_file" "source" {
  type        = "zip"
  source_dir  = local.root_dir
  output_path = "/tmp/function.zip"
}

# This bucket will host the zipped file.
resource "google_storage_bucket" "bucket" {
  name     = "${var.project_id}-${var.function_name}"
  location = var.region
}

# Add the zipped file to the bucket.
resource "google_storage_bucket_object" "zip" {
  # Use an MD5 here. If there's no changes to the source code, this won't change either.
  # We can avoid unnecessary redeployments by validating the code is unchanged, and forcing
  # a redeployment when it has!
  name   = "${data.archive_file.source.output_md5}.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.source.output_path
}

# The cloud function resource.
resource "google_cloudfunctions_function" "function" {
  available_memory_mb = var.available_memory_mb
  entry_point         = var.entry_point
  ingress_settings    = "ALLOW_ALL"

  name                  = var.function_name
  project               = var.project_id
  region                = var.region
  runtime               = var.runtime
  service_account_email = google_service_account.function-sa.email
  timeout               = var.timeout
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = "${data.archive_file.source.output_md5}.zip"
  max_instances         = 3000

  event_trigger {
    event_type = "google.pubsub.topic.publish"
    resource   = google_pubsub_topic.topic.id
  }

  environment_variables = {
    SLACK_WEBHOOK = var.slack_webhook
  }

}

# IAM Configuration. This allows unauthenticated, public access to the function.
# Change this if you require more control here.
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

# This is the service account in which the function will act as.
resource "google_service_account" "function-sa" {
  account_id   = "function-sa"
  description  = "Controls the workflow for the cloud pipeline"
  display_name = "function-sa"
  project      = var.project_id
}

resource "google_project_iam_member" "admin-account-iam" {
  project = var.project_id
  role    = "roles/cloudsql.admin"
  member  = "serviceAccount:${google_service_account.function-sa.email}"
}

resource "google_project_iam_member" "function-admin-iam" {
  project = var.project_id
  role    = "roles/cloudfunctions.admin"
  member  = "serviceAccount:${google_service_account.function-sa.email}"
}

resource "google_project_iam_member" "act_as_cloudfunctions_developer" {
  project = var.project_id
  role    = "roles/cloudfunctions.developer"
  member  = "serviceAccount:${data.google_project.project.number}@cloudbuild.gserviceaccount.com"
}
