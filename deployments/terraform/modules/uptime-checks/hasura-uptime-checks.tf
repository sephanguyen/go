data "google_monitoring_notification_channel" "slack" {
  project      = var.project_id
  type         = "slack"
  display_name = "Monitoring channel"
}

resource "google_monitoring_uptime_check_config" "hasura_uptime_checks" {
  count            = length(var.hasura_paths)
  display_name     = "Uptime check for ${var.hasura_host}${var.hasura_paths[count.index]}"
  timeout          = "30s"
  period           = "300s"
  project          = var.project_id
  selected_regions = ["ASIA_PACIFIC", "EUROPE", "USA_VIRGINIA"]

  http_check {
    path         = var.hasura_paths[count.index]
    port         = var.hasura_port
    use_ssl      = true
    validate_ssl = true
  }

  content_matchers {
    content = "OK"
    matcher = "CONTAINS_STRING"
  }

  monitored_resource {
    type = "uptime_url"
    labels = {
      project_id = var.project_id
      host       = var.hasura_host
    }
  }

}

resource "google_monitoring_alert_policy" "hasura_down" {
  count   = length(var.hasura_paths)
  project = var.project_id

  display_name = "Hasura at https://${var.hasura_host}:${var.hasura_port}${var.hasura_paths[count.index]} uptime checks"
  combiner     = "OR"

  conditions {
    display_name = "Hasura at https://${var.hasura_host}:${var.hasura_port}${var.hasura_paths[count.index]} is down or metadata is inconsistent"
    condition_threshold {
      filter          = "metric.type=\"monitoring.googleapis.com/uptime_check/check_passed\" AND metric.label.check_id=\"${google_monitoring_uptime_check_config.hasura_uptime_checks[count.index].uptime_check_id}\" AND resource.type=\"uptime_url\""
      comparison      = "COMPARISON_GT"
      threshold_value = 1
      duration        = "0s"
      aggregations {
        alignment_period     = "1200s"
        per_series_aligner   = "ALIGN_NEXT_OLDER"
        cross_series_reducer = "REDUCE_COUNT_FALSE"
        group_by_fields = [
          "resource.*"
        ]
      }
    }
  }
  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}
