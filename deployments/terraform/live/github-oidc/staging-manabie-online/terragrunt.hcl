include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/github-oidc"
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals

  project_id = "staging-manabie-online"
  sa_name    = "stag-deploy-bot"
}

inputs = {
  project_id = local.project_id
  pool_id    = local.env.pool_id
  deploy_bot = {
    iam = {
      name       = local.sa_name
      project_id = local.project_id
      roles = [
        "roles/cloudbuild.builds.editor",
        "roles/cloudkms.cryptoKeyEncrypterDecrypter",
        "roles/cloudsql.client",
        "roles/container.admin",
        "roles/logging.admin",
        "roles/iam.serviceAccountUser",
      ]
    }
    service_account_id = "projects/${local.project_id}/serviceAccounts/${local.sa_name}@${local.project_id}.iam.gserviceaccount.com"
    wif = {
      attribute_condition = local.env.deploy_bot_attribute_condition
    }
  }
  mfe_deploy_bot = {
    iam = {
      bucket_names = [
        "import-map-deployer-staging"
      ]
      name       = "stag-mfe-upload-artifacts"
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
        attribute.workflow == 'mfe.build' ||
        attribute.workflow == 'tbd.deploy' ||
        attribute.workflow == 'mfe.deploy'
      EOT
    }
  }
  get_release_tag_bot = {
    iam = {
      name       = "stag-get-release-tag-bot"
      project_id = local.project_id
      roles = [
        "roles/iam.workloadIdentityUser",
        "roles/container.viewer",
      ]
    }
    service_account_id = "projects/${local.project_id}/serviceAccounts/stag-get-release-tag@${local.project_id}.iam.gserviceaccount.com"
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'tbd.get_current_release' ||
        attribute.workflow == 'tbd.create_release_tag' || 
        attribute.workflow == 'tbd.build_production_with_uat_tags.yml' ||
        attribute.workflow == 'tbd.pick-tag-and-deploy-uat' ||
        attribute.workflow == 'tbd.deploy'
      EOT
    }
  }
}
