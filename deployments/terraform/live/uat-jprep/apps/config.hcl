locals {
  runs_on_project_id = "staging-manabie-online"

  org = "jprep"

  service_account_prefix = "uat-"

  kms_key_name = "uat-jprep"

  gke_rbac_enabled = true

  platforms_gke_dep        = "../../stag-manabie/platforms2"
  platforms_kms_dep        = "../../stag-manabie/platforms"
  platforms_storage_dep    = "../../stag-manabie/platforms"
  platforms_postgresql_dep = "../platforms"
}
