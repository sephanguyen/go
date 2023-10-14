include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps"
}

dependency "tokyo-platforms" {
  config_path = "../platforms"
}

dependency "platforms" {
  config_path = "../analytics-platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = true

  postgresql = {
    project_id = dependency.platforms.outputs.postgresql_project
    instance   = dependency.platforms.outputs.postgresql_instance
    databases = [
      "redash",
    ]
    users = []
  }

  # this is using a different database instance so we must set it
  # to empty so it won't run the postgresql_role_grant resource
  postgresql_bypass_rls_roles  = []
  postgresql_replication_roles = []

  service_accounts = [
    {
      name    = "${include.env.locals.service_account_prefix}redash"
      project = include.env.locals.project_id
      roles = {
        "${include.env.locals.project_id}" = [
          "roles/cloudsql.client",
          "roles/cloudsql.instanceUser",
          "roles/alloydb.client"
        ],
        "production-renseikai" = [
          "roles/cloudsql.client",
          "roles/cloudsql.instanceUser",
        ],
        "synersia" = [
          "roles/cloudsql.client",
          "roles/cloudsql.instanceUser",
        ],
        "production-manabie-vn" = [
          "roles/cloudsql.client",
          "roles/cloudsql.instanceUser",
          "roles/alloydb.client"
        ],
        "live-manabie" = [
          "roles/cloudsql.client",
          "roles/cloudsql.instanceUser",
          "roles/alloydb.client"
        ],
      }
      identity_namespaces = [
        include.env.locals.identity_namespaces.services,
      ]
    },
  ]

  gke_endpoint           = dependency.tokyo-platforms.outputs.gke_endpoint
  gke_ca_cert            = dependency.tokyo-platforms.outputs.gke_ca_cert
  gke_identity_namespace = try(dependency.tokyo-platforms.outputs.gke_identity_namespace, "")

  gke_rbac = {
    enabled = true
    policies = [
      {
        kind      = "Group"
        group     = "tech-func-platform@manabie.com"
        role_kind = "ClusterRole"
        role_name = "cluster-admin"
        namespaces = [
          "prod-${include.env.locals.org}-services",
        ]
      },
    ]
  }

  kms_keys = {
    "${include.env.locals.kms_key_name}" = merge(
      include.env.locals.kms_key_props,
      {
        key_ring        = dependency.tokyo-platforms.outputs.kms_key_ring
        rotation_period = "7776000s"
        decrypters = concat(
          [
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}redash"
            }
          ]
        )
      },
    )
  }
}
