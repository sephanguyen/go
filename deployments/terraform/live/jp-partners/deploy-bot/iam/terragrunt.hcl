include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../../modules//deploy-bot/iam"
}

dependency "service_account" {
  config_path = "../service-account"

  mock_outputs = {
    dorp_deploy_bot_service_account = {
      id    = "projects/student-coach-e1e95/serviceAccounts/dorp-deploy-bot@dummy.mock.email"
      email = "dorp-deploy-bot@dummy.mock.email"
    }
  }
}

inputs = {
  project_id                      = "student-coach-e1e95"
  dorp_deploy_bot_service_account = dependency.service_account.outputs.dorp_deploy_bot_service_account
  production_databases = [
    {
      project_id  = "student-coach-e1e95"
      instance_id = "manabie-2db8"
    },
    {
      project_id  = "student-coach-e1e95"
      instance_id = "jp-partners-b04fbb69"
    }
  ]
  preproduction_databases = [
    {
      project_id  = "student-coach-e1e95"
      instance_id = "clone-manabie-2db8"
    },
    {
      project_id  = "student-coach-e1e95"
      instance_id = "clone-jp-partners-b04fbb69"
    },
    {
      project_id  = "student-coach-e1e95"
      instance_id = "clone-jprep-6a98"
    }
  ]
}