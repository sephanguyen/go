include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps/v2"
}

dependency "platforms" {
  config_path = "../../stag-manabie/platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/uat-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = false

  service_accounts = concat(
    include.env.locals.service_accounts,
  )

}
