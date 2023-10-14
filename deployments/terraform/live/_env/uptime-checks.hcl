locals {
  config   = read_terragrunt_config("config.hcl", {})
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
}

inputs = {
  project_id = try(
    local.config.locals.runs_on_project_id,
    local.env_vars.locals.project_id,
  )
  hasura_host  = local.config.locals.hasura_host
  hasura_port  = try(local.config.locals.hasura_port, "443")
  hasura_paths = local.config.locals.hasura_paths
  https_check = try(local.config.locals.https_check, {})
}
