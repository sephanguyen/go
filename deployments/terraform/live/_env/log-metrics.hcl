locals {
  config   = read_terragrunt_config("config.hcl", {})
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
}

dependency "platforms" {
  config_path = "../platforms"
}

inputs = {
  project_id = try(
    local.config.locals.runs_on_project_id,
    local.env_vars.locals.project_id,
  )

  gke_cluster_name = dependency.platforms.outputs.gke_cluster_name
}
