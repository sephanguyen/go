include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/project-roles"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/project-roles.hcl"
}
