include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/logging-project-exclusions"
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals

  project_id = "staging-manabie-online"
}

inputs = {
  project_id = local.project_id
  exclusions = [
    {
      name   = "gce_instance"
      filter = <<EOT
resource.type="gce_instance"
EOT
    },
    {
      name   = "istio"
      filter = <<EOT
resource.type="k8s_container"
AND resource.labels.namespace_name="istio-system" 
AND severity=INFO
EOT
    },
    {
      name   = "k8s_containers"
      filter = <<EOT
resource.type="k8s_container" 
AND severity<"WARNING"
EOT
    },
  ]
}
