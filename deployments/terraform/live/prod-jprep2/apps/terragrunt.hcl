include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps"
}

dependency "platforms" {
  config_path = "../platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = true

  service_accounts = [
    for sa in include.env.locals.service_accounts : {
      name    = sa.name
      project = sa.project
      roles = merge(
        sa.roles,
        sa.name == "${include.env.locals.service_account_prefix}tom" ? {
          # Since we are using a single Identity Platform resource on student-coach-e1e95
          # project for all partners, the service account for "tom" service need to be granted the
          # Firebase Cloud Messaging Admin role so it can use the FCM to send notifications
          # for all partners. See https://manabie.atlassian.net/browse/LT-13515.
          student-coach-e1e95 = concat(
            lookup(sa.roles, "student-coach-e1e95", []),
            [
              "roles/firebasenotifications.admin",
              "roles/firebasecloudmessaging.admin",
            ],
          )
        } : {},
        sa.name == "${include.env.locals.service_account_prefix}yasuo" ? {
          # Since we are using a single Identity Platform resource on student-coach-e1e95
          # project for all partners, the service account for "yasuo" service need to be granted the
          # Firebase Cloud Messaging Admin role so it can use the FCM to send notifications
          # for all partners. See https://manabie.atlassian.net/browse/LT-13515.
          student-coach-e1e95 = concat(
            lookup(sa.roles, "student-coach-e1e95", []),
            [
              "roles/firebasenotifications.admin",
              "roles/firebasecloudmessaging.admin",
            ],
          )
        } : {},

        # We are still using live-manabie Firebase project to send notifications for JPREP, so
        # the `notificationmgmt` service account need to be granted the Firebase Cloud Messaging
        # role on that project.
        sa.name == "${include.env.locals.service_account_prefix}notificationmgmt" ? {
          live-manabie = concat(
            lookup(sa.roles, "live-manabie", []),
            [
              "roles/firebasenotifications.admin",
              "roles/firebasecloudmessaging.admin",
            ],
          )
        } : {},
      )
      identity_namespaces = sa.identity_namespaces
      bucket_roles        = try(sa.bucket_roles, {})
      impersonations      = try(sa.impersonations, null)
    }
    if sa.name != "${include.env.locals.service_account_prefix}entryexitmgmt-hasura"
    # temporary ignore entryexitmgmt-hasura service account for now, since its
    # fully qualified service account name exceeds the limitation of 30 characters,
    # and it's not deployed for JPREP yet.
  ]

  gke_backup_plan = [
    {
      # gke_cluster_id format:
      #     projects/student-coach-e1e95/locations/asia-northeast1/clusters/tokyo
      project  = split("/", dependency.gke.outputs.gke_cluster_id)[1]
      cluster  = dependency.gke.outputs.gke_cluster_id
      name     = "prod-jprep-kafka"
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
            namespace = "prod-jprep-kafka"
            name      = "kafka"
          }
        ]
      }
    }
  ]

  kms_keys = {
    for key, props in include.env.inputs.kms_keys :
    key => merge(
      props,
      {
        decrypters = [
          for dec in props.decrypters : dec
          if dec.service_account_name != "${include.env.locals.service_account_prefix}entryexitmgmt-hasura"
          # temporary ignore entryexitmgmt-hasura service account for now, since its
          # fully qualified service account name exceeds the limitation of 30 characters,
          # and it's not deployed for JPREP yet.
        ]
      },
    )
  }
}
