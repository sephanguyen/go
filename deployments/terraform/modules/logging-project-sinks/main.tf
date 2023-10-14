resource "google_logging_project_sink" "project-sink" {
  for_each = { for sink in var.sinks : "${sink.name}" => sink }

  name        = each.value.name
  project     = var.project_id
  destination = each.value.destination
  description = each.value.description
  filter      = each.value.filter

  dynamic "exclusions" {
    for_each = { for e in each.value.exclusions : "${e.name}" => e }

    content {
      name        = exclusions.value["name"]
      filter      = exclusions.value["filter"]
      description = exclusions.value["description"]
    }
  }

  disabled               = each.value.disabled
  unique_writer_identity = each.value.unique_writer_identity
}
