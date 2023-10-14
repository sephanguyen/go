locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  project_id = local.env_vars.locals.project_id

  config                = read_terragrunt_config("config.hcl", {})
  platforms_dep         = try(local.config.locals.platforms_dep, "../platforms")
  postgresql_project_id = try(local.config.locals.postgresql_project_id, local.project_id)

  postgresql_port_config = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/postgresql-port.hcl")

  roles = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/roles.hcl")
}

dependency "access_level" {
  config_path = "../../workspace/access-control"
}

dependency "platforms" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_dep}"
}

inputs = {
  project_id = local.project_id

  postgresql_project_id    = local.postgresql_project_id
  postgresql_instance      = dependency.platforms.outputs.postgresql_instance
  postgresql_instance_port = local.postgresql_port_config.locals.postgresql_instance_port

  role_by_access_level   = local.roles.locals.role_by_access_level
  member_by_access_level = dependency.access_level.outputs.members_by_access_level
}
