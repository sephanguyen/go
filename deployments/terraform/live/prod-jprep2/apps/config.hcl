locals {
  runs_on_project_id = "student-coach-e1e95"

  org = "jprep"

  service_account_prefix = "prod-jprep-"

  gke_rbac_enabled = true

  kms_key_name            = "prod-jprep"
  kms_key_rotation_period = "7776000s"

  # Deploy JPREP to Tokyo GKE cluster.
  platforms_gke = "../../prod-tokyo/platforms"

  # Reuse the Cloud Storage bucket in the old project for now, since
  # changing the bucket name would require a lot of work, i.e. updating
  # all media records in the database to use the new bucket name.
  platforms_storage = "../../prod-jprep/platforms"
}
