include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/cloud-build-trigger"
}

locals {
  stag = read_terragrunt_config("stag_triggers.hcl").locals
  uat  = read_terragrunt_config("uat_triggers.hcl").locals
  dorp = read_terragrunt_config("dorp_triggers.hcl").locals
  prod = read_terragrunt_config("prod_triggers.hcl").locals
}

inputs = {
  ad_hoc_infos = concat(
    local.stag.ad_hoc_infos,
    local.uat.ad_hoc_infos,
    local.dorp.ad_hoc_infos,
    local.prod.ad_hoc_infos,
  )
}
