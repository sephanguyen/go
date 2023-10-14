locals {
  project_id = "staging-manabie-online"
  region     = "asia-southeast1"

  env = "stag"

  # the bucket where Terraform state will be stored
  bucket = "manabie-staging-terraform-state"
}
