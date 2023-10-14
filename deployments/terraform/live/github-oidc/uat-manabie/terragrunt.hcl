include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/github-oidc"
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals

  project_id = "uat-manabie"
}

inputs = {
  project_id = local.project_id
  pool_id    = local.env.pool_id
  mfe_deploy_bot = {
    iam = {
      bucket_names = [
        "import-map-deployer-uat",
        "import-map-deployer-preproduction"
      ]
      name       = "uat-mfe-upload-artifacts"
      project_id = local.project_id
      roles = {
        "student-coach-e1e95" = [
          "roles/artifactregistry.writer"
        ],
        "${local.project_id}" = []
      }
    }
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'tbd.build' ||
        attribute.workflow == 'tbd.deploy' ||
        attribute.workflow == 'mfe.build' ||
        attribute.workflow == 'mfe.deploy'
      EOT
    }
  }
}