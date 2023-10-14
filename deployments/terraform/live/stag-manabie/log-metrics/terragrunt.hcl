include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/log-metrics"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/log-metrics.hcl"
}
