include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = false

  gke_backup_plan = [
    {
      # gke_cluster_id format:
      #     projects/student-coach-e1e95/locations/asia-northeast1/clusters/jp-partners
      project  = split("/", dependency.gke.outputs.gke_cluster_id)[1]
      cluster  = dependency.gke.outputs.gke_cluster_id
      name     = "prod-renseikai-kafka"
      location = "asia-northeast1"
      retention_policy = {
        backup_delete_lock_days = 5
        backup_retain_days      = 5
      }
      cron_schedule = "0 * * * *"
      backup_config = {
        include_volume_data = true
        include_secrets     = true
        selected_applications = [
          {
            namespace = "prod-renseikai-kafka"
            name      = "kafka"
          }
        ]
      }
    }
  ]
}
