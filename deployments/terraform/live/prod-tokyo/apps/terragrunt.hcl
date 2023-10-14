include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps/v2"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = true

  service_accounts = concat(
    [
      for sa in include.env.locals.service_accounts : {
        name    = sa.name
        project = sa.project
        roles = merge(
          sa.roles,
          sa.name == "${include.env.locals.service_account_prefix}usermgmt" ? {
            student-coach-e1e95 = concat(
              lookup(sa.roles, "student-coach-e1e95", []),
              [
                "roles/identityplatform.admin"
              ],
            )
          } : {}
        )
        identity_namespaces = sa.identity_namespaces
        bucket_roles        = try(sa.bucket_roles, {})
        impersonations      = try(sa.impersonations, null)
      }
    ],
    [
      {
        name    = "grafana"
        project = include.env.locals.project_id
        roles = {
          "${include.env.locals.project_id}" = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/monitoring.viewer",
          ],
          "live-manabie" = [
            "roles/monitoring.viewer",
          ],
          "staging-manabie-online" = [
            "roles/monitoring.viewer",
          ]
        }
        identity_namespaces = [
          "monitoring",
        ]
      },
    ],

    # for data-warehouse
    [
      {
        name    = "${include.env.locals.service_account_prefix}dwh-cp-schema-registry"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
          "dorp-${include.env.locals.org}-data-warehouse",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}dwh-kafka"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
          "dorp-${include.env.locals.org}-data-warehouse",
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
          "${include.env.locals.env}-${include.env.locals.org}-kafka",
          "${include.env.locals.env}-${include.env.locals.org}-services",
          "${include.env.locals.env}-${include.env.locals.org}-backend",
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
          "dorp-${include.env.locals.org}-kafka",
          "dorp-${include.env.locals.org}-services",
          "dorp-${include.env.locals.org}-backend",
          "dorp-${include.env.locals.org}-data-warehouse",
        ]
      },
      {
        name    = "${include.env.locals.service_account_prefix}dwh-cp-ksql-server"
        project = include.env.locals.project_id
        roles   = {}
        identity_namespaces = [
          "${include.env.locals.env}-${include.env.locals.org}-data-warehouse",
          "dorp-${include.env.locals.org}-data-warehouse",
        ]
      },
    ],
  )

  gke_backup_plan = [
    {
      # gke_cluster_id format:
      #     projects/student-coach-e1e95/locations/asia-northeast1/clusters/tokyo
      project  = split("/", dependency.gke.outputs.gke_cluster_id)[1]
      cluster  = dependency.gke.outputs.gke_cluster_id
      name     = "prod-tokyo-kafka"
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
            namespace = "prod-tokyo-kafka"
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
        # for data-warehouse
        decrypters = concat(
          props.decrypters,
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
