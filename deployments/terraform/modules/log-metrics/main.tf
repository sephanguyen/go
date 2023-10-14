data "google_monitoring_notification_channel" "slack" {
  project      = var.project_id
  type         = "slack"
  display_name = "Monitoring channel"
}

resource "google_logging_metric" "istio_nr_host" {
  project = var.project_id
  name    = "${var.gke_cluster_name}-istio_nr_host"
  filter  = <<-EOT
    resource.type="k8s_container"
    resource.labels.cluster_name="${var.gke_cluster_name}"
    resource.labels.namespace_name="istio-system"
    resource.labels.container_name="istio-proxy"
    jsonPayload.response_code="404"
    jsonPayload.response_flags="NR"
    jsonPayload.path=~"^\/manabie\..+|.+\.v\d\..+"
    jsonPayload.method="OPTIONS"
  EOT

  metric_descriptor {
    metric_kind = "DELTA"
    value_type  = "INT64"
  }
}

resource "google_monitoring_alert_policy" "istio_nr_host" {
  project = var.project_id

  display_name = "[${var.gke_cluster_name}] Istio proxy got NR response flags"
  combiner     = "OR"

  conditions {
    display_name = "Istio proxy got NR (no route to host) response flags, likely the Gateway or Virtualservice are misconfigured"
    condition_threshold {
      filter = "metric.type=\"logging.googleapis.com/user/${google_logging_metric.istio_nr_host.id}\" resource.type=\"k8s_container\""

      duration        = "0s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_COUNT"
        cross_series_reducer = "REDUCE_COUNT"
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

resource "google_monitoring_alert_policy" "prometheus_restarting" {
  project = var.project_id

  display_name = "[${var.gke_cluster_name}] Prometheus server is crash looping"
  combiner     = "OR"

  conditions {
    display_name = "Prometheus server is crash looping"

    condition_threshold {
      filter = <<-EOT
        metric.type="kubernetes.io/container/restart_count"
        resource.type="k8s_container"
        resource.label."container_name"="prometheus-server"
        resource.label."cluster_name"="${var.gke_cluster_name}"
      EOT

      duration        = "900s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0

      aggregations {
        alignment_period   = "900s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

resource "google_monitoring_alert_policy" "alertmanager_restarting" {
  project = var.project_id

  display_name = "[${var.gke_cluster_name}] Alertmanager is crash looping"
  combiner     = "OR"

  conditions {
    display_name = "Alertmanager is crash looping"

    condition_threshold {
      filter = <<-EOT
        metric.type="kubernetes.io/container/restart_count"
        resource.type="k8s_container"
        resource.label."container_name"="prometheus-alertmanager"
        resource.label."cluster_name"="${var.gke_cluster_name}"
      EOT

      duration        = "900s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0

      aggregations {
        alignment_period   = "900s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}
