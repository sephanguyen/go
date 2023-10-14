module "simple_bucket" {
  source  = "terraform-google-modules/cloud-storage/google//modules/simple_bucket"
  version = "4.0.0"

  project_id  = var.project_id
  name        = var.bucket_name
  location    = var.bucket_location
  iam_members = var.bucket_iam_members

  public_access_prevention = var.bucket_public_access_prevention
}

variable "project_id" {
  type        = string
  description = "The ID of the project to create the bucket in."
}

variable "bucket_name" {
  type        = string
  description = "The name of the bucket."
}

variable "bucket_location" {
  type        = string
  description = "The location of the bucket."
}

variable "bucket_iam_members" {
  description = "The list of IAM members to grant permissions on the bucket."
  type = list(object({
    role   = string
    member = string
  }))
  default = []
}

variable "bucket_public_access_prevention" {
  description = "Prevents public access to a bucket. Acceptable values are inherited or enforced. If inherited, the bucket uses public access prevention, only if the bucket is subject to the public access prevention organization policy constraint."
  type        = string
  default     = "inherited"
}
