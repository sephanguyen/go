include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/github-oidc"
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals

  project_id = "student-coach-e1e95"
  sa_name    = "prod-deploy-bot"
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
        "roles/cloudsql.admin",
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
  build_bot = {
    iam = {
      name       = "prod-build-bot"
      project_id = local.project_id
      roles = {
        "${local.project_id}" = [
          "roles/storage.objectViewer", # to pull Docker blobs from Container Registry
          "roles/artifactregistry.writer",
        ]
      }
    }
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'build.runner' ||
        attribute.workflow == 'mfe.build'    ||
        attribute.workflow == 'tbd.build-v2' ||
        attribute.workflow == 'tbd.build' ||
        attribute.workflow == 'build.j4' ||
        attribute.workflow == 'tbd.build-and-deploy-learnosity'
      EOT
    }
  }
  integration_test_bot = {
    iam = {
      name       = "integration-test-bot"
      project_id = local.project_id
      roles = {
        "${local.project_id}" = [
          "roles/storage.objectViewer",    # to pull Docker blobs from Container Registry
          "roles/artifactregistry.reader", # to pull Docker blobs from Artifact Registry
        ]
      }
    }
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'tiered.pre_merge' ||
        attribute.workflow == 'tiered.post_merge_integration_test' ||
        attribute.workflow == 'tiered.regression'
      EOT
    }
  }
  unleash_decryptor_bot = {
    iam = {
      name       = "unleash-decryptor-bot"
      project_id = local.project_id
      roles      = {}
    }
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'tbd.unleash' ||
        attribute.workflow == 'unleash_health_check'
      EOT
    }
  }
  dorp_deploy_bot = {
    service_account_id = "projects/${local.project_id}/serviceAccounts/dorp-deploy-bot@${local.project_id}.iam.gserviceaccount.com"
    wif = {
      attribute_condition = local.env.deploy_bot_attribute_condition
    }
  }

  mfe_deploy_bot = {
    iam = {
      bucket_names = [
        "import-map-deployer-production"
      ]
      name       = "prod-mfe-upload-artifacts"
      project_id = local.project_id
      roles = {
        "${local.project_id}" = [
          "roles/artifactregistry.writer"
        ]
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
  get_release_tag_bot = {
    iam = {
      name       = "prod-get-release-tag-bot"
      project_id = local.project_id
      roles = [
        "roles/iam.workloadIdentityUser",
        "roles/container.viewer",
      ]
    }
    service_account_id = "projects/${local.project_id}/serviceAccounts/prod-get-release-tag@${local.project_id}.iam.gserviceaccount.com"
    wif = {
      attribute_condition = <<EOT
        attribute.workflow == 'tbd.get_current_release' ||
        attribute.workflow == 'tbd.create_release_tag'|| 
        attribute.workflow == 'tbd.build_production_with_uat_tags.yml' ||
        attribute.workflow == 'tbd.pick-tag-and-deploy-uat' ||
        attribute.workflow == 'tbd.deploy'
      EOT
    }
  }
}
