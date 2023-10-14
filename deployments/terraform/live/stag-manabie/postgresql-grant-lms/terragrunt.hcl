include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql-grant"
}

dependency "postgresql" {
  config_path = "../postgresql-lms"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/postgresql-grant-v2.hcl"
}

inputs = {
  postgresql_instance = "manabie-lms-de12e08e"
  postgresql_databases = [
    for db in dependency.postgresql.outputs.postgresql_databases : db
    if db == "eureka" || db == "uat_eureka"
  ]
}
