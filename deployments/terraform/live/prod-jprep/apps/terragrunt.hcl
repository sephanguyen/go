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
          # project for all partners, this tom service account need to be granted the
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
          # project for all partners, this yasuo service account need to be granted the
          # Firebase Cloud Messaging Admin role so it can use the FCM to send notifications
          # for all partners. See https://manabie.atlassian.net/browse/LT-13515.
          student-coach-e1e95 = concat(
            lookup(sa.roles, "student-coach-e1e95", []),
            [
              "roles/firebasenotifications.admin",
              "roles/firebasecloudmessaging.admin",
            ],
          )
        } : {}
      )
      identity_namespaces = sa.identity_namespaces
      bucket_roles        = try(sa.bucket_roles, {})
      impersonations      = try(sa.impersonations, null)
    }
  ]
}
