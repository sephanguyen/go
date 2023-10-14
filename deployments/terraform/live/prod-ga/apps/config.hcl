locals {
  runs_on_project_id = "student-coach-e1e95"

  org = "ga"

  db_prefix      = "ga_"
  db_user_prefix = "ga_"

  service_account_prefix = "prod-"

  kms_key_name = "prod-ga"

  platforms_gke        = "../../jp-partners/platforms"
  platforms_kms        = "../../jp-partners/platforms"
  platforms_storage    = "../../jp-partners/platforms"
  platforms_postgresql = "../../jp-partners/platforms"
}
