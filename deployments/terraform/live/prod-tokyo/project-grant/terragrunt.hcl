include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/project-grant"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/project-grant.hcl"
}
