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
  path   = "${get_terragrunt_dir()}/../../_env/stag-apps.hcl"
  expose = true
}

inputs = {
  pgaudit_enabled = true

  postgresql = merge(
    include.env.inputs.postgresql,
    {
      instance = dependency.platforms.outputs.postgresql_instances.common
      databases = concat(
        [
          # don't need to create `eureka` and `auth` database, since they
          # will be created in LMS and Auth SQL instance, respectively
          for db in include.env.inputs.postgresql.databases : db
          if db != "${include.env.locals.db_prefix}eureka" && db != "${include.env.locals.db_prefix}auth"
        ],
        [
          "${include.env.locals.db_prefix}redash",
        ],
        [
          "${include.env.locals.db_prefix}alloydb",
        ],
        [
          "${include.env.locals.db_prefix}kec",
        ],
      )
      users = concat(
        # TODO: do we need to create eureka's service account user here?
        include.env.locals.users,
        [
          "redash",
          "${include.env.locals.service_account_prefix}redash@${include.env.locals.project_id}.iam",

          # for data-warehouse
          "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
        ]
      )
    }
  )

  postgresql_user_permissions = flatten(concat(
    [
      # don't need to grant permissions for `eureka` and `shamir` databases,
      # since they are put in LMS and Auth SQL instances, respectively
      for p in include.env.locals.postgresql_user_permissions : p
      if p.database != "${include.env.locals.db_prefix}eureka" && p.database != "${include.env.locals.db_prefix}auth"
    ],
    [
      {
        database    = "${include.env.locals.db_prefix}bob"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}bob"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}tom"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}tom"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}fatima"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}fatima"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}invoicemgmt"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}invoicemgmt"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}draft"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}draft"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "${include.env.locals.service_account_prefix}kafka-connect@${include.env.locals.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "${include.env.locals.service_account_prefix}kafka-connect@${include.env.locals.project_id}.iam"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE", "CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE", "CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "redash"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}alloydb"
        user        = "redash"
        schema      = "bob"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "${include.env.locals.service_account_prefix}kafka-connect@${include.env.locals.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "${include.env.locals.service_account_prefix}kafka-connect@${include.env.locals.project_id}.iam"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE", "CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE", "CREATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "redash"
        schema      = "bob"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}kec"
        user        = "redash"
        schema      = "bob"
        object_type = "table"
        privileges  = ["SELECT"]
      },
    ]
  ))

  postgresql_bypass_rls_roles = flatten(concat(
    include.env.inputs.postgresql_bypass_rls_roles,

    # for data-warehouse
    [
      "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
    ],
  ))
}
