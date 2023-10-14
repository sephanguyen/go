variable "project_id" {
  description = "The project ID in Google Cloud to use for these resources."
  type        = string
}

variable "region" {
  description = "The region in Google Cloud where the resources will be deployed."
  default     = "asia-southeast1"
  type        = string
}

variable "function_name" {
  description = "The name of the function to be deployed"
  type        = string
}

variable "entry_point" {
  description = "The entrypoint where the function is called"
  type        = string
}

variable "available_memory_mb" {
  description = "Request memory for the function"
  default     = "128"
  type        = string
}

variable "runtime" {
  description = "The function runtime"
  default     = "go120"
  type        = string
}

variable "timeout" {
  description = "The function timeout"
  default     = 180
}

variable "slack_webhook" {
  description = "Slack webhook for sending notifications"
  type        = string
}

variable "topic_name" {
  description = "Cloud Pub/Sub topic name"
  type        = string
}

variable "scheduler_jobs" {
  description = "List of scheduler jobs"
  type        = list(object({ name = string, description = string, schedule = string, data = string }))
}
