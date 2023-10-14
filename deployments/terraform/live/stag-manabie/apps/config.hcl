locals {
  runs_on_project_id = "staging-manabie-online"

  org = "manabie"

  service_account_prefix = "stag-"

  gke_rbac_enabled = true

  kms_key_name = "github-actions"

  platforms_gke_dep = "../platforms2"
}
