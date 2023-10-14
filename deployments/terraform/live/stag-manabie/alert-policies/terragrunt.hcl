include {
  path = find_in_parent_folders()
}

locals {
  env    = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  common = read_terragrunt_config(find_in_parent_folders("common.hcl")).locals
}

terraform {
  source = "../../../modules/alert-policies"
}

inputs = {
  project_id = local.env.project
  region     = local.env.region

  hasura_metric_name = local.common.hasura_metric_name
  k8s_cluster_name   = local.env.k8s_cluster_name
}
