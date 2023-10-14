locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  project_id = local.env_vars.locals.project_id
  env        = local.env_vars.locals.env

  roles = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/roles.hcl")
}

dependency "access_level" {
  config_path = "../../workspace/access-control"
}

inputs = {
  project_id = local.project_id
  env        = local.env_vars.locals.env

  role_by_access_level   = local.roles.locals.role_by_access_level
  member_by_access_level = dependency.access_level.outputs.members_by_access_level

  techleads      = dependency.access_level.outputs.techleads
  techlead_roles = local.roles.locals.techlead_roles
}
