include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/logging-project-exclusions"
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals

  project_id = "student-coach-e1e95"
}

inputs = {
  project_id = local.project_id
  exclusions = [
    {
      name   = "manabie-b2c"
      filter = <<EOT
resource.labels.namespace_name="old-prod-manabie-backend" OR 
resource.labels.namespace_name="old-prod-manabie-nats-streaming" 
OR ( 
resource.type="k8s_container" 
AND resource.labels.namespace_name="istio-system" 
AND severity=INFO
)
EOT
    },
    {
      name   = "all_prod"
      filter = <<EOT
resource.type="k8s_container" 
AND resource.labels.namespace_name=~"(prod|dorp).+(services|backend)"
AND severity<"WARNING"
EOT
    },
    {
      name   = "unleash"
      filter = <<EOT
resource.type="k8s_container" 
AND resource.labels.namespace_name=~".+unleash"
AND severity<"WARNING"
EOT
    },
  ]
}
