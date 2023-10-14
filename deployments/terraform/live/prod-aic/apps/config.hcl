locals {
  runs_on_project_id = "student-coach-e1e95"

  org = "aic"

  db_prefix      = "aic_"
  db_user_prefix = "aic_"

  service_account_prefix = "prod-"

  kms_key_name = "prod-aic"

  platforms_gke        = "../../jp-partners/platforms"
  platforms_kms        = "../../jp-partners/platforms"
  platforms_storage    = "../../jp-partners/platforms"
  platforms_postgresql = "../../jp-partners/platforms"
}
