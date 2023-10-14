resource "google_monitoring_notification_channel" "monitoring_alerts_tf" {
  type         = "slack"
  display_name = "Monitoring Alerts (tf)"
  project      = var.project_id

  labels = {
    "channel_name" = var.slack_channel
  }
  sensitive_labels {
    auth_token = var.slack_auth_token
  }
}