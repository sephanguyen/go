include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/vpc"
}

inputs = {
  network_name            = "production"
  auto_create_subnetworks = true
  subnets                 = []
  secondary_ranges        = {}
}
