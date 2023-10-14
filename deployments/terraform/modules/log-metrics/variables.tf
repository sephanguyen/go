variable "project_id" {
  description = "The project ID to host the cluster in"
  type        = string
}

variable "gke_cluster_name" {
  type        = string
  description = "Kubernetes cluster name"
}
