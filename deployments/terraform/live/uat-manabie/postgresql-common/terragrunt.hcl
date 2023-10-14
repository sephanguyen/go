include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql"
}

dependency "platforms" {
  config_path = "../../stag-manabie/platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/uat-apps.hcl"
  expose = true
}

inputs = {
  pgaudit_enabled = true

  postgresql = merge(
    include.env.inputs.postgresql,
    {
      instance = dependency.platforms.outputs.postgresql_instances.common
      databases = [
        # don't need to create eureka database,
        # since that database will be created in different SQL instance
        for db in include.env.inputs.postgresql.databases : db
        if db != "${include.env.locals.db_prefix}eureka" && db != "${include.env.locals.db_prefix}auth"
      ]

      # TODO: do we need to create eureka's service account user here?
      users = concat(
        include.env.inputs.postgresql.users,
        [
          # Note that this account is imported, NOT created by this project.
          # When deleting this user state, run `tg state rm` instead of actually deleting
          # the resource.
          # for now only UAT have data of KEC.
          # but DWH system only have in staging. To save, we decide to point DWH staging sync data from UAT
          "stag-kafka-connect@staging-manabie-online.iam",
        ],
      )
    }
  )

  postgresql_user_permissions = concat(
    [
      # don't need to grant permissions for eureka database,
      # since that database is in different SQL instance
      for p in include.env.locals.postgresql_user_permissions : p
      if p.database != "${include.env.locals.db_prefix}eureka" && p.database != "${include.env.locals.db_prefix}auth"
    ],

    [
      {
        database    = "${include.env.locals.db_prefix}bob"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}bob-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}bob"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}bob-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}timesheet"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}timesheet-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}timesheet"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}timesheet-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}invoicemgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}invoicemgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}invoicemgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}invoicemgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}fatima"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}fatima-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}fatima"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}fatima-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}mastermgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}mastermgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}mastermgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}mastermgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}calendar"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}calendar-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}calendar"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}calendar-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
      {
        database    = "${include.env.locals.db_prefix}lessonmgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}lessonmgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${include.env.locals.db_prefix}lessonmgmt"
        user        = "stag-kafka-connect@staging-manabie-online.iam"
        owner       = "${include.env.locals.service_account_prefix}lessonmgmt-m@${include.env.locals.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE"]
      },
    ],
  )
}
