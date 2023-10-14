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
      instance = dependency.platforms.outputs.postgresql_instances.auth
      databases = [
        "${include.env.locals.db_prefix}auth",
      ]
      users = concat(
        # TODO: need to figure out which users are needed for auth, then add them here
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
      for p in include.env.locals.postgresql_user_permissions : p
      if p.database == "${include.env.locals.db_prefix}auth"
    ],
  ))

  postgresql_bypass_rls_roles = flatten(concat(
    include.env.inputs.postgresql_bypass_rls_roles,

    # for data-warehouse
    [
      "${include.env.locals.service_account_prefix}dwh-kafka-connect@${include.env.locals.project_id}.iam",
    ],
  ))
}
