include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/project-services"
}

inputs = {
  activate_apis = [
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "cloudkms.googleapis.com",
    "sqladmin.googleapis.com",
    "servicenetworking.googleapis.com",
    "vpcaccess.googleapis.com",
  ]
}
