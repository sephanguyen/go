resource "google_pubsub_topic" "topic" {
  name = var.topic_name
}

resource "google_cloud_scheduler_job" "job" {

  for_each = {
    for index, job in var.scheduler_jobs :
    job.name => job
  }

  name        = each.value.name
  description = each.value.description
  schedule    = each.value.schedule
  time_zone   = "Asia/Ho_Chi_Minh"

  pubsub_target {
    topic_name = google_pubsub_topic.topic.id
    data       = base64encode(each.value.data)
  }
}
