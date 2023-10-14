include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/apps"
}

dependency "postgresql" {
  config_path = "../platforms"
}

include "env" {
  path   = "${get_terragrunt_dir()}/../../_env/uat-apps.hcl"
  expose = true
}

inputs = {
  create_storage_hmac_key = false
}
