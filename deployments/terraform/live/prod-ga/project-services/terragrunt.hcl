include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/project-services"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/project-services.hcl"
}
