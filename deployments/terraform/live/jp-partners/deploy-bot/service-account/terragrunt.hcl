include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../../modules//deploy-bot/service-account"
}

locals {
  token = yamldecode(sops_decrypt_file(find_in_parent_folders("token.enc.yaml")))
}

inputs = {
  project_id   = "student-coach-e1e95"
  github_token = local.token.github
}
