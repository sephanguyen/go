module "bigquery" {
  count   = var.bigquery.enabled ? 1 : 0
  source  = "terraform-google-modules/bigquery/google"
  version = "5.4.0"

  dataset_id                 = var.bigquery.dataset_id
  dataset_name               = var.bigquery.dataset_name
  description                = var.bigquery.description
  project_id                 = var.project_id
  location                   = var.bigquery.location
  delete_contents_on_destroy = var.bigquery.delete_contents_on_destroy
  dataset_labels             = var.bigquery.dataset_labels
}
