locals {
  transformed_https_check = flatten([
    for item in var.https_check : [
      for path in item.paths : {
        host                           = item.host
        port                           = item.port
        name                           = "${item.name} https://${item.host}:${item.port}${path} uptime check"
        path                           = path
        content_matchers               = item.content_matchers
        accepted_response_status_codes = item.accepted_response_status_codes
        request_method                 = item.request_method
        content_type                   = item.content_type
        body                           = item.body
      }
    ]
  ])
}

resource "google_monitoring_uptime_check_config" "https_check_updtime_checks" {
  count = length(local.transformed_https_check)

  display_name     = local.transformed_https_check[count.index].name
  timeout          = "60s"
  period           = "300s" # 5 minutes
  project          = var.project_id
  selected_regions = ["ASIA_PACIFIC", "EUROPE", "USA_VIRGINIA"]

  http_check {
    path           = local.transformed_https_check[count.index].path
    port           = local.transformed_https_check[count.index].port
    use_ssl        = true
    validate_ssl   = true
    request_method = local.transformed_https_check[count.index].request_method
    content_type   = local.transformed_https_check[count.index].content_type
    body           = local.transformed_https_check[count.index].body
    accepted_response_status_codes {
      status_value = local.transformed_https_check[count.index].accepted_response_status_codes.status_value
      status_class = local.transformed_https_check[count.index].accepted_response_status_codes.status_class
    }
  }

  content_matchers {
    content = local.transformed_https_check[count.index].content_matchers.content
    matcher = local.transformed_https_check[count.index].content_matchers.matcher
  }

  monitored_resource {
    type = "uptime_url"
    labels = {
      project_id = var.project_id
      host       = local.transformed_https_check[count.index].host
    }
  }

}

resource "google_monitoring_alert_policy" "https_check_down" {
  count = length(local.transformed_https_check)

  project = var.project_id

  display_name = local.transformed_https_check[count.index].name
  combiner     = "OR"

  conditions {
    display_name = "${local.transformed_https_check[count.index].name} is down"
    condition_threshold {
      filter          = "metric.type=\"monitoring.googleapis.com/uptime_check/check_passed\" AND metric.label.check_id=\"${google_monitoring_uptime_check_config.https_check_updtime_checks[count.index].uptime_check_id}\" AND resource.type=\"uptime_url\""
      comparison      = "COMPARISON_GT"
      threshold_value = 1
      duration        = "60s"
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
