locals {
  runs_on_project_id = "staging-manabie-online"

  org = "jprep"

  db_prefix      = "stag_"
  db_user_prefix = "stag_"

  service_account_prefix = "stag-jprep-"

  kms_key_name = "stag-jprep"

  gke_rbac_enabled = true

  platforms_gke_dep        = "../../stag-manabie/platforms2"
  platforms_kms_dep        = "../../stag-manabie/platforms"
  platforms_storage_dep    = "../../stag-manabie/platforms"
  platforms_postgresql_dep = "../../uat-jprep/platforms"
}
