include {
  path = find_in_parent_folders()
}

locals {
  env    = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  common = read_terragrunt_config(find_in_parent_folders("common.hcl")).locals
}

terraform {
  source = "../../../modules/notification-channel"
}

inputs = {
  project_id = local.env.project
  region     = local.env.region

  slack_channel    = "#test-monitoring"
  slack_auth_token = local.common.slack_auth_token
}
