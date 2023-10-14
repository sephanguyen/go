resource "google_monitoring_alert_policy" "hasura_slow" {
  project = var.project_id

  display_name = "Hasura slow query execution time (tf)"
  combiner     = "OR"

  conditions {
    display_name = "logging/user/${var.hasura_metric_name} [95TH PERCENTILE]"
    condition_threshold {
      filter          = "metric.type=\"logging.googleapis.com/user/${var.hasura_metric_name}\" resource.type=\"k8s_container\""
      duration        = "900s"
      comparison      = "COMPARISON_GT"
      threshold_value = 1.5
      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_DELTA"
        cross_series_reducer = "REDUCE_PERCENTILE_95"
        group_by_fields      = []
      }
      trigger {
        count = 1
      }
    }
  }
}
