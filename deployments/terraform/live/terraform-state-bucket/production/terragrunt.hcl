include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/simple-bucket"
}

inputs = {
  project_id      = "student-coach-e1e95"
  bucket_name     = "manabie-terraform-state-2"
  bucket_location = "ASIA-SOUTHEAST1"
  bucket_iam_members = [
    {
      role   = "roles/storage.objectAdmin"
      member = "user:bao.nguyen@manabie.com"
    },
    {
      role   = "roles/storage.objectAdmin"
      member = "user:tuananh.pham@manabie.com"
    },
  ]
  bucket_public_access_prevention = "enforced"
}
