resource "google_logging_project_exclusion" "project-exclusion" {
  count = length(var.exclusions)
  name = var.exclusions[count.index].name
  description = var.exclusions[count.index].description
  filter = var.exclusions[count.index].filter
  project = var.project_id
}
