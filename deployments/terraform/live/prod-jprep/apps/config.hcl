locals {
  runs_on_project_id = "live-manabie"

  org = "jprep"

  service_account_prefix = "prod-"

  gke_rbac_enabled = true

  kms_key_name            = "prod-jprep"
  kms_key_rotation_period = "7776000s"
}
