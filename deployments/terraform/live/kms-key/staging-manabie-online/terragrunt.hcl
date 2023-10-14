include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/kms-key"
}

locals {
  project_id               = "staging-manabie-online"
  stag_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/stag-defs.yaml"))
  uat_service_definitions  = yamldecode(file("${get_repo_root()}/deployments/decl/uat-defs.yaml"))

  dwh_custom_services = [
    "dwh-cp-schema-registry",
    "dwh-kafka",
    "dwh-kafka-connect",
    "dwh-cp-ksql-server",
  ]

  kms_keys = merge(
    {
      for service in local.stag_service_definitions :
      "stag-${service.name}" => {
        rotation_period     = "2592000s"
        owner               = "group:dev.infras@manabie.com"
        encrypter_decrypter = "stag-${service.name}-techlead@manabie.com"
        encrypters          = []
        decrypters = distinct(compact([
          "stag-${service.name}@${local.project_id}.iam.gserviceaccount.com",
          "stag-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "stag-${service.name}-h@${local.project_id}.iam.gserviceaccount.com" : "",
          "stag-jprep-${service.name}@${local.project_id}.iam.gserviceaccount.com",
          "stag-jprep-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "stag-jprep-${service.name}-h@${local.project_id}.iam.gserviceaccount.com" : "",
        ]))
      } if !lookup(service, "disable_iam", false)
    },
    {
      for service_name in local.dwh_custom_services :
      "stag-${service_name}" => {
        rotation_period     = "2592000s"
        owner               = "group:dev.infras@manabie.com"
        encrypter_decrypter = "stag-${service_name}-techlead@manabie.com"
        encrypters          = []
        decrypters = [
          "stag-${service_name}@${local.project_id}.iam.gserviceaccount.com",
          "stag-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          "stag-jprep-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
        ]
      }
    },
    {
      for service in local.uat_service_definitions :
      "uat-${service.name}" => {
        rotation_period     = "2592000s"
        owner               = "group:dev.infras@manabie.com"
        encrypter_decrypter = "uat-${service.name}-techlead@manabie.com"
        encrypters          = []
        decrypters = distinct(compact([
          "uat-${service.name}@uat-manabie.iam.gserviceaccount.com",
          "uat-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "uat-${service.name}-h@uat-manabie.iam.gserviceaccount.com" : "",
          "uat-${service.name}@${local.project_id}.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "uat-${service.name}-h@${local.project_id}.iam.gserviceaccount.com" : "",
        ]))
      } if !lookup(service, "disable_iam", false)
    }
  )
}

inputs = {
  project_id = local.project_id
  key_ring = {
    name     = "backend-services"
    location = "global"
  }
  kms_keys             = local.kms_keys
  create_google_groups = true
}
