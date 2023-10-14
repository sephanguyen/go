include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/kms-key"
}

locals {
  project_id               = "student-coach-e1e95"
  prod_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/prod-defs.yaml"))

  dwh_custom_services = [
    "dwh-cp-schema-registry",
    "dwh-kafka",
    "dwh-kafka-connect",
    "dwh-cp-ksql-server",
  ]

  kms_keys = merge(
    {
      for service in local.prod_service_definitions :
      "prod-${service.name}" => {
        rotation_period     = "2592000s"
        owner               = "group:dev.infras@manabie.com"
        encrypter_decrypter = "prod-${service.name}-techlead@manabie.com"
        encrypters          = []
        decrypters = distinct(compact([
          "prod-${service.name}@production-aic.iam.gserviceaccount.com",
          "prod-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          "prod-jprep-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-aic.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-ga.iam.gserviceaccount.com",
          "prod-nats-jetstream@synersia.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-renseikai.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-${service.name}-h@production-aic.iam.gserviceaccount.com" : "",
          "prod-${service.name}@production-ga.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-${service.name}-h@production-ga.iam.gserviceaccount.com" : "",
          "prod-${service.name}@production-renseikai.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-${service.name}-h@production-renseikai.iam.gserviceaccount.com" : "",
          "prod-${service.name}@synersia.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-${service.name}-h@synersia.iam.gserviceaccount.com" : "",
          "prod-${service.name}@student-coach-e1e95.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-${service.name}-h@student-coach-e1e95.iam.gserviceaccount.com" : "",

          # These service accounts are used for JPREP deployments in student-coach-e1e95 project
          "prod-jprep-${service.name}@student-coach-e1e95.iam.gserviceaccount.com",
          try(service.hasura.v2_enabled, false) || try(service.hasura.enabled, false) ? "prod-jprep-${service.name}-h@student-coach-e1e95.iam.gserviceaccount.com" : "",
        ]))
      } if !lookup(service, "disable_iam", false)
    },
    {
      for service_name in local.dwh_custom_services :
      "prod-${service_name}" => {
        rotation_period     = "2592000s"
        owner               = "group:dev.infras@manabie.com"
        encrypter_decrypter = "prod-${service_name}-techlead@manabie.com"
        encrypters          = []
        decrypters = [
          "prod-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          "prod-jprep-nats-jetstream@${local.project_id}.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-aic.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-ga.iam.gserviceaccount.com",
          "prod-nats-jetstream@synersia.iam.gserviceaccount.com",
          "prod-nats-jetstream@production-renseikai.iam.gserviceaccount.com",
          "prod-${service_name}@student-coach-e1e95.iam.gserviceaccount.com",
        ]
      }
    },
  )
}

inputs = {
  project_id = local.project_id
  key_ring = {
    name     = "backend-services"
    location = "asia-northeast1"
  }
  kms_keys             = local.kms_keys
  create_google_groups = true
}
