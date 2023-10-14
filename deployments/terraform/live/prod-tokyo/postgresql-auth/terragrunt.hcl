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
      project_id             = "production-aic"
      service_account_prefix = "prod-"
    },
    {
      project_id             = "production-ga"
      service_account_prefix = "prod-"
    },
    {
      project_id             = "production-renseikai"
      service_account_prefix = "prod-"
    },
    {
      project_id             = "synersia"
      service_account_prefix = "prod-"
    },
  ]
}

inputs = {
  pgaudit_enabled = true

  postgresql = merge(
    include.env.inputs.postgresql,
    {
      instance = dependency.platforms.outputs.postgresql_instances.auth
      databases = [
        "${include.env.locals.db_prefix}auth",
      ]
      users = flatten(concat(
        # TODO: need to figure out which users are needed for auth, then add them here
        include.env.locals.users,
        [
          "grafana@${include.env.locals.project_id}.iam",

          # for data-warehouse
          "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
        ],

        # We create extra users so that pods services from aic/ga/renseikai/synersia
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

        # Hasura service accounts are a different story
        [
          for service in include.env.locals.service_definitions : [
            for extra in local.extra_orgs : "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
          ] if try(service.postgresql.createdb, false) && (try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false))
        ],
      ))
    }
  )

  postgresql_user_permissions = flatten(concat(
    [
      for p in include.env.locals.postgresql_user_permissions : p
      if p.database == "${include.env.locals.db_prefix}auth"
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
        ] if grant.dbname == "auth"
      ]
    ],

    # Skip for hasura but need to grant for hasura SA.
    [
      for service in include.env.locals.service_definitions : [
        for extra in local.extra_orgs : [
          {
            database    = "${include.env.locals.db_prefix}${service.name}"
            user        = "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
            schema      = "public"
            object_type = "table"
            privileges  = service.name != "draft" ? ["SELECT", "INSERT", "UPDATE"] : ["SELECT", "INSERT", "UPDATE", "DELETE"]
          },
          {
            database    = "${include.env.locals.db_prefix}${service.name}"
            user        = "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
            schema      = "public"
            object_type = "sequence"
            privileges  = ["USAGE", "SELECT", "UPDATE"]
          },
          {
            database    = "${include.env.locals.db_prefix}${service.name}"
            user        = "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
            schema      = "public"
            object_type = "function"
            privileges  = ["EXECUTE"]
          },
          {
            database    = "${include.env.locals.db_prefix}${service.name}"
            user        = "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
            schema      = "public"
            object_type = "schema"
            privileges  = ["USAGE"]
          },
        ]
      ] if(try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)) && service.name == "auth"
    ],
  ))

  postgresql_bypass_rls_roles = flatten(concat(
    include.env.inputs.postgresql_bypass_rls_roles,
    [
      for service in include.env.locals.service_definitions : [
        for extra in local.extra_orgs : "${extra.service_account_prefix}${service.name}@${extra.project_id}.iam"
      ] if try(service.postgresql.bypassrls, false)
    ],

    [
      for service in include.env.locals.service_definitions : [
        for extra in local.extra_orgs : "${extra.service_account_prefix}${service.name}-h@${extra.project_id}.iam"
      ] if try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)
    ],

    # for data-warehouse
    [
      "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
    ],
  ))
}
