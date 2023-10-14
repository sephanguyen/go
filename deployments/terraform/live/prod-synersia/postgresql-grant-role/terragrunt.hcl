include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql-grant-role"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/postgresql-grant-role.hcl"
}
