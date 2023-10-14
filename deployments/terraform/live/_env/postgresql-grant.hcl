locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  project_id = local.env_vars.locals.project_id
  env        = local.env_vars.locals.env

  config                = read_terragrunt_config("config.hcl", {})
  apps_dep              = try(local.config.locals.apps_dep, "../apps")
  platforms_dep         = try(local.config.locals.platforms_dep, "../platforms")
  postgresql_roles_dep  = try(local.config.locals.postgresql_roles_dep, "../postgresql-roles")
  postgresql_project_id = try(local.config.locals.postgresql_project_id, local.project_id)

  postgresql_port_config = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/postgresql-port.hcl")
}

dependency "postgresql-roles" {
  config_path = "${get_terragrunt_dir()}/${local.postgresql_roles_dep}"
}

dependency "platforms" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_dep}"
}

dependency "apps" {
  config_path = "${get_terragrunt_dir()}/${local.apps_dep}"
}

inputs = {
  project_id = local.project_id
  env        = local.env_vars.locals.env

  postgresql_project_id           = local.postgresql_project_id
  postgresql_instance             = dependency.platforms.outputs.postgresql_instance
  postgresql_instance_port        = local.postgresql_port_config.locals.postgresql_instance_port
  postgresql_databases            = dependency.apps.outputs.postgresql_databases
  postgresql_read_only_role_name  = dependency.postgresql-roles.outputs.postgresql_read_only_role_name
  postgresql_read_write_role_name = dependency.postgresql-roles.outputs.postgresql_read_write_role_name
}
