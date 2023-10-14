include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps/v2"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/stag-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = false

  gke_backup_plan = [
    {
      # gke_cluster_id format:
      #     projects/staging-manabie-online/locations/asia-southeast1-b/clusters/staging-2
      project  = split("/", dependency.platforms_gke.outputs.gke_cluster_id)[1]
      cluster  = dependency.platforms_gke.outputs.gke_cluster_id
      name     = "stag-jprep-kafka"
      location = "asia-southeast1"
      retention_policy = {
        backup_delete_lock_days = 3
        backup_retain_days      = 3
      }
      cron_schedule = "1 0 * * *"
      backup_config = {
        include_volume_data = true
        include_secrets     = true
        selected_applications = [
          {
            namespace = "stag-jprep-kafka"
            name      = "kafka"
          }
        ]
      }
    }
  ]
}
