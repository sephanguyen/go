module "kms" {
  count = var.kms.enabled ? 1 : 0

  source  = "terraform-google-modules/kms/google"
  version = "2.1.0"

  project_id = var.project_id
  location   = var.kms.location
  keyring    = var.kms.keyring
}
