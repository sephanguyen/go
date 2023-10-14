include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql"
}

dependency "platforms" {
  config_path = "../../uat-jprep/platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/stag-apps.hcl"
  expose = true
}

inputs = {
  postgresql = merge(
    include.env.inputs.postgresql,
    {
      instance = dependency.platforms.outputs.postgresql_instances.jprep
    }
  )
}
