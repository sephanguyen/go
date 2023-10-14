include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql-roles"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/postgresql-roles.hcl"
}

inputs = {
  postgresql_instance = "manabie-auth-f2dc7988"
}
