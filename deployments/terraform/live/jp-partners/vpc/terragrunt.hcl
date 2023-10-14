include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/vpc"
}

inputs = {
  project_id              = "student-coach-e1e95"
  network_name            = "jp-partners"
  auto_create_subnetworks = false
  subnets = [
    {
      subnet_name   = "jp-partners"
      subnet_region = "asia-northeast1"
      subnet_ip     = "10.146.0.0/20"
    }
  ]
  secondary_ranges = {
    jp-partners = [
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
