resource "google_cloudbuild_trigger" "adhoc" {
  for_each = {
    for p in var.ad_hoc_infos :
    "${p.org}.${p.env}.${p.service_account_adhoc}" => p
  }

  project         = each.value.project_id
  name            = "${each.value.env}-${each.value.org}-ad-hoc"
  description     = "Trigger to run ad-hoc SQL and K8S in ${each.value.env}-${each.value.org}. Managed by Terraform."
  disabled        = false
  service_account = "projects/${each.value.project_id}/serviceAccounts/${each.value.service_account_adhoc}"
  substitutions = {
    _ORG  = "${each.value.org}"
    _ENV  = "${each.value.env}"
  }
  source_to_build {
    uri       = "https://github.com/manabie-com/backend"
    ref       = "refs/heads/develop"
    repo_type = "GITHUB"
  }
  git_file_source {
    path      = "deployments/google_cloud_build/adhoc.yaml"
    uri       = "https://github.com/manabie-com/backend"
    revision  = "refs/heads/develop"
    repo_type = "GITHUB"
  }
  approval_config {
    approval_required = true
  }
}
