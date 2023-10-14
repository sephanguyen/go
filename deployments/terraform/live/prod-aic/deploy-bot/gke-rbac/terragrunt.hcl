include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../../modules//deploy-bot/gke-rbac"
}

dependency "jp_partners_service_account" {
  config_path = "../../../jp-partners/deploy-bot/service-account"

  mock_outputs = {
    dorp_deploy_bot_service_account = {
      id    = "projects/student-coach-e1e95/serviceAccounts/dorp-deploy-bot@dummy.mock.email"
      email = "dorp-deploy-bot@dummy.mock.email"
    }
  }
}

dependency "jp_partners_platforms" {
  config_path = "${get_terragrunt_dir()}/../../../jp-partners/platforms"
}

inputs = {
  project_id                      = "student-coach-e1e95"
  org                             = "aic"
  gke_endpoint                    = dependency.jp_partners_platforms.outputs.gke_endpoint
  gke_ca_cert                     = dependency.jp_partners_platforms.outputs.gke_ca_cert
  dorp_deploy_bot_service_account = dependency.jp_partners_service_account.outputs.dorp_deploy_bot_service_account
  configure_common_namespaces     = false
  configure_dorp_namespaces       = true
}
