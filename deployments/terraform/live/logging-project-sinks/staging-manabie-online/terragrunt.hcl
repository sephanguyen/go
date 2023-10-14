include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/logging-project-sinks"
}

locals {
  env        = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  project_id = "staging-manabie-online"
  orgs       = ["manabie", "jprep"]

  stag_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/stag-defs.yaml"))
  uat_service_definitions  = yamldecode(file("${get_repo_root()}/deployments/decl/uat-defs.yaml"))

  merge_service_definitions = flatten([
    flatten([for org in local.orgs : [
      for service in local.stag_service_definitions : {
        name      = "${service.name}"
        sink_name = "stag-${org}-${service.name}"
        namespace = service.name == "kafka-connect" ? "stag-${org}-kafka" : service.name == "elasticsearch" ? "stag-${org}-elastic" : "stag-${org}-(services|backend)"
    } if try(contains(service.identity_namespaces, "services"), false)]]),
    flatten([for org in local.orgs : [
      for service in local.uat_service_definitions : {
        name      = "${service.name}"
        sink_name = "uat-${org}-${service.name}"
        namespace = service.name == "kafka-connect" ? "uat-${org}-kafka" : service.name == "elasticsearch" ? "uat-${org}-elastic" : "uat-${org}-(services|backend)"
    } if try(contains(service.identity_namespaces, "services"), false)]])
    ]
  )
}

inputs = {
  project_id = local.project_id
  sinks = concat(
    [
      for service in local.merge_service_definitions : {
        name                   = "${service.sink_name}"
        destination            = "logging.googleapis.com/projects/${local.project_id}/locations/global/buckets/_Default"
        filter                 = <<EOT
                resource.type="k8s_container" 
                AND resource.labels.namespace_name=~"${service.namespace}"
                AND labels.k8s-pod/app_kubernetes_io/name="${service.name}"
                AND severity>=DEBUG
                AND timestamp<="2023-01-01T00:00:00Z"
                AND timestamp>="2023-01-01T00:00:00Z"
            EOT
        exclusions             = []
        unique_writer_identity = true
      }
    ],

    [
      # This records "VerifyAppVersion requests" log.
      # See https://manabie.slack.com/archives/C025EN333K8/p1686720188998349
      # It's disabled by default.
      {
        name                   = "check-VerifyAppVersion-request-payload"
        destination            = "logging.googleapis.com/projects/${local.project_id}/locations/global/buckets/_Default"
        filter                 = <<EOT
resource.type="k8s_container" 
AND resource.labels.namespace_name=~".+(services|backend)"
AND jsonPayload.msg="VerifyAppVersion request"
EOT
        exclusions             = []
        unique_writer_identity = true
        disabled               = true
      },
    ],
  )
}
