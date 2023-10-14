include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql"
}

dependency "platforms" {
  config_path = "../platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
  expose = true
}

locals {
  extra_orgs = [
    {
      project_id             = "synersia"
      service_account_prefix = "prod-"
    },
  ]
}

inputs = {
  pgaudit_enabled = false

  postgresql = merge(
    include.env.inputs.postgresql,
    {
      instance = dependency.platforms.outputs.postgresql_instances.tokyo
      databases = concat(
        [
          # don't need to create `eureka` and `auth` database,
          # since that database will be created in LMS SQL instance
          for db in include.env.inputs.postgresql.databases : db
          if db != "${include.env.locals.db_prefix}eureka" && db != "${include.env.locals.db_prefix}auth"
        ],
        [
          "grafana",
        ]
      )
      users = flatten(concat(
        # TODO: do we need to create eureka's service account user here?
        include.env.locals.users,
        [
          "grafana@${include.env.locals.project_id}.iam",

          # for data-warehouse
          "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
        ],

        # We create extra users so that pods services from synersia
        # can still connect to prod-tokyo. The logic is copied from prod-apps.hcl.
        # hasura is using user/pass so we ignore them.
        # redash and ad-hoc are not required.
        # See https://manabie.atlassian.net/browse/LT-29678
        [
          for service in include.env.locals.service_definitions : [
            # Note: this requires the other projects to be applied first,
            # so that the service account exists for these db user creations.
            for extra in local.extra_orgs : "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
          ] if !try(service.disable_iam, false)
        ],
      ))
    }
  )

  postgresql_user_permissions = flatten(concat(
    [
      # don't need to grant permissions for eureka database,
      # since that database is put in LMS SQL instance
      for p in include.env.locals.postgresql_user_permissions : p
      if p.database != "${include.env.locals.db_prefix}eureka" && p.database != "${include.env.locals.db_prefix}auth"
    ],

    # Grant services from other orgs access to prod-tokyo db, similar to prod-apps.hcl
    # Note that we skip granting for hasura, as we can use tokyo_hasura directly with password.
    [
      for service in include.env.locals.service_definitions : [
        for grant in try(service.postgresql.grants, []) : [
          for extra in local.extra_orgs : [
            {
              database    = "${include.env.locals.db_prefix}${grant.dbname}"
              user        = "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
              schema      = "public"
              object_type = "table"
              privileges  = lookup(grant, "grant_delete", false) ? ["SELECT", "INSERT", "UPDATE", "DELETE"] : ["SELECT", "INSERT", "UPDATE"]
            },
            {
              database    = "${include.env.locals.db_prefix}${grant.dbname}"
              user        = "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
              schema      = "public"
              object_type = "sequence"
              privileges  = ["USAGE", "SELECT", "UPDATE"]
            },
            {
              database    = "${include.env.locals.db_prefix}${grant.dbname}"
              user        = "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
              schema      = "public"
              object_type = "function"
              privileges  = ["EXECUTE"]
            },
            {
              database    = "${include.env.locals.db_prefix}${grant.dbname}"
              user        = "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
              schema      = "public"
              object_type = "schema"
              privileges  = ["USAGE"]
            },
          ]
        ] if grant.dbname != "eureka" && grant.dbname != "auth"
      ]
    ],
    [
      for extra in local.extra_orgs : [
        {
          database    = "${include.env.locals.db_prefix}bob"
          user        = "${extra.service_account_prefix}shamir@${extra.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT"]
          objects = [
            "users",
            "users_groups",
            "teachers",
            "school_admins",
            "students",
            "parents",
            "organization_auths",
            "api_keypair",
          ]
        },
      ]
    ]
  ))

  # Grant BYPASSRLS for database accounts from other partners, like synersia.
  postgresql_bypass_rls_roles = flatten(concat(
    include.env.inputs.postgresql_bypass_rls_roles,
    [
      for service in include.env.locals.service_definitions : [
        for extra in local.extra_orgs : "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
      ] if try(service.postgresql.bypassrls, false)
    ],

    # for data-warehouse
    [
      "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
    ],
  ))
}
