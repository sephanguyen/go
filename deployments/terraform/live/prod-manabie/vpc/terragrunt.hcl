include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/vpc"
}

inputs = {
  project_id              = "student-coach-e1e95"
  network_name            = "manabie"
  auto_create_subnetworks = false
  subnets = [
    {
      subnet_name   = "manabie"
      subnet_region = "asia-southeast1"
      subnet_ip     = "10.146.0.0/20"
    }
  ]
  secondary_ranges = {
    manabie = [
      {
        range_name    = "gke-range-pods"
        ip_cidr_range = "10.32.0.0/14"
      },
      {
        range_name    = "gke-range-services"
        ip_cidr_range = "10.97.0.0/20"
      }
    ]
  }
}
