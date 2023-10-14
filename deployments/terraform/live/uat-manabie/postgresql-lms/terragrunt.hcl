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
      instance = dependency.platforms.outputs.postgresql_instances.lms
      databases = [
        "${include.env.locals.db_prefix}eureka",
      ]
      # TODO: need to figure out which users are needed for eureka, then add them here
    }
  )

  postgresql_user_permissions = [
    for p in include.env.locals.postgresql_user_permissions : p
    if p.database == "${include.env.locals.db_prefix}eureka"
  ]
}
