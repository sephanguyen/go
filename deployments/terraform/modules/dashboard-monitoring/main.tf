locals {
  json_file = fileset(path.module, "default/*.json")
}

resource "google_monitoring_dashboard" "default_dashboards" {
  for_each       = local.json_file
  project        = var.project_id
  dashboard_json = file(each.key)
}
