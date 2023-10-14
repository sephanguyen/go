module "vpc" {
  source  = "terraform-google-modules/network/google"
  version = "3.3.0"

  project_id   = var.project_id
  network_name = var.network_name

  auto_create_subnetworks = var.auto_create_subnetworks

  subnets          = var.subnets
  secondary_ranges = var.secondary_ranges
}

resource "google_compute_global_address" "google-managed-services-range" {
  project       = var.project_id
  name          = "google-managed-services-${module.vpc.network_name}"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 24
  network       = module.vpc.network_name
}

resource "google_service_networking_connection" "private_service_access" {
  network                 = module.vpc.network_name
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.google-managed-services-range.name]
}
