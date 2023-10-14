include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../../modules//deploy-bot/iam"
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

inputs = {
  project_id                      = "synersia"
  dorp_deploy_bot_service_account = dependency.jp_partners_service_account.outputs.dorp_deploy_bot_service_account
  production_databases = [
    {
      project_id  = "synersia"
      instance_id = "synersia-228d"
    },
  ]
  preproduction_databases = [
    {
      project_id  = "synersia"
      instance_id = "clone-synersia-228d"
    },
  ]
}
