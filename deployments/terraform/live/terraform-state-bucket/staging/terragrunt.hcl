include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/simple-bucket"
}

inputs = {
  project_id      = "student-coach-e1e95"
  bucket_name     = "manabie-staging-terraform-state"
  bucket_location = "ASIA-SOUTHEAST1"
  bucket_iam_members = [
    {
      role   = "roles/storage.objectViewer"
      member = "user:huubang.nguyen@manabie.com"
    },
    {
      role   = "roles/storage.objectViewer"
      member = "user:chivy.nguyen@manabie.com"
    },
  ]
  bucket_public_access_prevention = "enforced"
}
