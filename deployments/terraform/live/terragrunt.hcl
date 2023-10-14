locals {
  # Load any configured variables
  env    = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  common = read_terragrunt_config(find_in_parent_folders("common.hcl")).locals

  default_bucket = "manabie-terraform-state-2" # -2 suffix because manabie-terraform-state bucket name already exists
  bucket         = try(local.env.bucket, local.default_bucket)
}

# Configure Terragrunt to store state in GCP buckets
remote_state {
  backend = "gcs"

  config = {
    bucket = "${local.bucket}"
    prefix = "${path_relative_to_include()}/"
  }

  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
}

generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "google" {
  project     = "${local.env.project_id}"
  region      = "${local.env.region}"
}
EOF
}

# The generated file is named `_versions_override.tf`, so that it
# comes first in lexicographical order, thus applied first before
# any other overrides are applied.
generate "versions" {
  path      = "_versions_override.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
terraform {
  required_version = "~> 1.0"
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "3.90.0"
    }
    google-beta = {
      source = "hashicorp/google-beta"
      version = "3.90.0"
    }
  }
}
EOF
}

inputs = merge(
  local.env,
  local.common,
)
