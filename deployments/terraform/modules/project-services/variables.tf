variable "project_id" {
  type = string
}

variable "activate_apis" {
  type = list(string)

  default = [
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "cloudkms.googleapis.com",
    "sqladmin.googleapis.com",
    "servicenetworking.googleapis.com",
  ]
}
