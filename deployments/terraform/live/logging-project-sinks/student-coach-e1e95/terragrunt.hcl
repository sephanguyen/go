include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/logging-project-sinks"
}

locals {
  env                      = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  project_id               = "student-coach-e1e95"
  orgs                     = ["jprep", "synersia", "renseikai", "ga", "aic", "tokyo"]
  prod_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/prod-defs.yaml"))
  merge_service_definitions = flatten([for org in local.orgs : [
    for service in local.prod_service_definitions : {
      name      = "${service.name}"
      sink_name = "prod-${org}-${service.name}"
      namespace = "prod-${org}-services"
  } if try(contains(service.identity_namespaces, "services"), false)]])
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
                AND resource.labels.namespace_name="${service.namespace}"
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
      # Enable this project sink when you need to check whether the old ports
      # are still being requested
      {
        name                   = "check-deprecated-port-LT-42684"
        destination            = "logging.googleapis.com/projects/${local.project_id}/locations/global/buckets/_Default"
        filter                 = <<EOT
    resource.type="k8s_container"
    AND resource.labels.namespace_name="istio-system"
    AND resource.labels.container_name="istio-proxy"
    AND jsonPayload.authority=~".*manabie.io:.*"
    AND jsonPayload.authority!="thanos-sidecar.jp-partners.manabie.io:443"
    EOT
        exclusions             = []
        disabled               = false
        unique_writer_identity = true
      },

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
AND jsonPayload."grpc.request.header.authority"=~".*31400|31500|31600.*"
EOT
        exclusions             = []
        unique_writer_identity = true
        disabled               = false
      },
    ],
  )
}
