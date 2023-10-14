locals {
  runs_on_project_id = "student-coach-e1e95"

  org = "synersia"

  service_account_prefix = "prod-"

  kms_key_name = "prod-synersia"

  platforms_gke        = "../../jp-partners/platforms"
  platforms_kms        = "../../jp-partners/platforms"
  platforms_storage    = "../../jp-partners/platforms"
  platforms_postgresql = "../platforms"
}
