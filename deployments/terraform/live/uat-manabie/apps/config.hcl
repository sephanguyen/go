locals {
  runs_on_project_id = "staging-manabie-online"

  org = "manabie"

  db_prefix      = "uat_"
  db_user_prefix = "uat_"

  service_account_prefix = "uat-"

  gke_rbac_enabled = true

  kms_key_name = "uat-manabie"

  platforms_gke_dep        = "../../stag-manabie/platforms2"
  platforms_kms_dep        = "../../stag-manabie/platforms"
  platforms_storage_dep    = "../../stag-manabie/platforms"
  platforms_postgresql_dep = "../../stag-manabie/platforms"
}
