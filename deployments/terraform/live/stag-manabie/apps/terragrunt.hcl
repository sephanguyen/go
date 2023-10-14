include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps/v2"
}

dependency "platforms" {
  config_path = "../platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/stag-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = true

  service_accounts = concat(
    include.env.locals.service_accounts,
    [
      {
        name    = "${include.env.locals.service_account_prefix}redash"
        project = include.env.locals.project_id
        roles = {
          staging-manabie-online = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/alloydb.client"
          ]
        }
        identity_namespaces = [
          "redash",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}aphelios"
        project = include.env.locals.project_id
        roles = {
          staging-manabie-online = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
        }
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-machine-learning",
        ]
        bucket_roles = {
          stag-manabie-backend = [
            "roles/storage.legacyBucketWriter"
          ]
        }
      },
      {
        name    = "${include.env.locals.service_account_prefix}pachyderm"
        project = include.env.locals.project_id
        roles = {
          staging-manabie-online = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
        }
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-machine-learning",
        ]
        bucket_roles = {
          stag-manabie-backend = [
            "roles/storage.legacyBucketWriter",
          ]
        }
      },
      {
        name    = "${include.env.locals.service_account_prefix}mlflow"
        project = include.env.locals.project_id
        roles = {
          staging-manabie-online = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
        }
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-machine-learning",
        ]
        bucket_roles = {
          stag-manabie-backend = [
            "roles/storage.legacyBucketWriter"
          ]
        }
      },
      {
        name    = "${include.env.locals.service_account_prefix}kserve"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-machine-learning",
        ]
        bucket_roles = {
          stag-manabie-backend = [
            "roles/storage.legacyBucketWriter"
          ]
        }
      },
      {
        name    = "${include.env.locals.service_account_prefix}identity-hook-runner"
        project = include.env.locals.project_id
        roles = {
          staging-manabie-online = [
            "roles/firebaseauth.admin",
            "roles/identityplatform.admin",
            "roles/storage.admin",
            "roles/iam.serviceAccountAdmin",
            "roles/iam.workloadIdentityUser",
          ]
        }
        identity_namespaces = []
      }
    ],

    # for data-warehouse
    [
      {
        name    = "${include.env.locals.service_account_prefix}dwh-cp-schema-registry"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}dwh-kafka"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}dwh-kafka-connect"
        project = include.env.locals.project_id
        roles = {
          "${include.env.locals.project_id}" = [
            "roles/alloydb.client",
            "roles/bigquery.dataOwner",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
        }
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-services",
          "${include.env.locals.env}-${include.env.locals.org}-backend",
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}dwh-cp-ksql-server"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
        ]
      },
    ],
  )

  bucket_name = "stag-manabie-backend"

  kms_keys = {
    "${include.env.locals.kms_key_name}" = merge(
      include.env.locals.kms_key_props,
      {
        key_ring        = dependency.platforms.outputs.kms_key_ring
        rotation_period = "7776000s"
        decrypters = concat(
          include.env.locals.kms_key_props.decrypters,
          [
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}redash"
            },
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}aphelios"
            },
          ],

          # for data-warehouse
          [
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}dwh-cp-schema-registry"
            },
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}dwh-kafka"
            },
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}dwh-kafka-connect"
            },
            {
              service_account_project = include.env.locals.project_id
              service_account_name    = "${include.env.locals.service_account_prefix}dwh-cp-ksql-server"
            },
          ],
        )
      },
    )
  }
}
