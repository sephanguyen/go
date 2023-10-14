include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql-grant-role"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/postgresql-grant-role.hcl"
}

inputs = {
  bypass_rls_role_write_privileges_enabled = true
  postgresql_use_predefined_roles          = true
}
