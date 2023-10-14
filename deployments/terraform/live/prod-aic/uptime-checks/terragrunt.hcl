include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/uptime-checks"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/uptime-checks.hcl"
}