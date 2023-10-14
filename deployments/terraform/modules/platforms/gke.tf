module "gke" {
  count = var.gke.enabled ? 1 : 0

  source  = "terraform-google-modules/kubernetes-engine/google//modules/beta-public-cluster"
  version = "25.0.0"

  project_id = var.project_id
  name       = var.gke.cluster_name
  region     = var.gke.region
  regional   = var.gke.regional
  zones      = var.gke.zones

  kubernetes_version = var.gke.kubernetes_version
  release_channel    = var.gke.release_channel

  enable_vertical_pod_autoscaling = true
  authenticator_security_group    = var.gke.security_group

  network           = var.gke.network_name
  subnetwork        = var.gke.subnetwork_name
  ip_range_pods     = var.gke.ip_range_pods
  ip_range_services = var.gke.ip_range_services

  cluster_telemetry_type   = "ENABLED"
  remove_default_node_pool = true
  grant_registry_access    = true
  gce_pd_csi_driver        = var.gke.gce_pd_csi_driver

  service_account        = var.gke.service_account
  create_service_account = var.gke.create_service_account

  cluster_autoscaling = {
    enabled             = var.gke.cluster_autoscaling.enabled
    autoscaling_profile = var.gke.cluster_autoscaling.autoscaling_profile
    min_cpu_cores       = var.gke.cluster_autoscaling.min_cpu_cores
    max_cpu_cores       = var.gke.cluster_autoscaling.max_cpu_cores
    min_memory_gb       = var.gke.cluster_autoscaling.min_memory_gb
    max_memory_gb       = var.gke.cluster_autoscaling.max_memory_gb
    gpu_resources       = var.gke.cluster_autoscaling.gpu_resources
    auto_repair         = var.gke.cluster_autoscaling.auto_repair
    auto_upgrade        = var.gke.cluster_autoscaling.auto_upgrade
  }

  node_pools = [
    for np in var.gke.node_pools : {
      name           = np.name
      machine_type   = np.machine_type
      autoscaling    = np.autoscaling
      node_count     = np.node_count
      min_count      = np.min_count
      max_count      = np.max_count
      image_type     = np.image_type
      spot           = np.spot
      disk_size_gb   = np.disk_size_gb
      disk_type      = np.disk_type
      node_locations = np.node_locations
      auto_upgrade   = true
      auto_repair    = true
    }
  ]

  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  node_pools_labels          = var.gke.node_pools_labels
  node_pools_metadata        = var.gke.node_pools_metadata
  node_pools_taints          = var.gke.node_pools_taints
  node_pools_tags            = var.gke.node_pools_tags
  node_pools_resource_labels = var.gke.node_pools_resource_labels

  network_policy = var.gke.network_policy

  maintenance_start_time = var.gke.maintenance_start_time
  maintenance_end_time   = var.gke.maintenance_end_time
  maintenance_recurrence = var.gke.maintenance_recurrence

  gke_backup_agent_config = var.gke.backup_agent_config
}

locals {
  namespaces = var.gke_rbac.enabled ? distinct(flatten(var.gke_rbac.policies[*].namespaces)) : []

  per_namespace_policies = var.gke_rbac.enabled ? flatten([
    for policy in var.gke_rbac.policies : [
      for ns in policy.namespaces : {
        kind      = policy.kind
        group     = policy.group
        namespace = ns
        role_kind = policy.role_kind
        role_name = policy.role_name
      }
    ]
  ]) : []

  cluster_wide_policies = var.gke_rbac.enabled ? flatten([
    for policy in var.gke_rbac.policies : {
      kind      = policy.kind
      group     = policy.group
      role_kind = policy.role_kind
      role_name = policy.role_name
    } if length(policy.namespaces) == 0
  ]) : []
}

resource "kubernetes_namespace" "namespaces" {
  for_each = toset(local.namespaces)

  metadata {
    name = each.value
  }

  lifecycle {
    ignore_changes = [
      metadata,
    ]
  }
}

resource "kubernetes_role_binding" "role_binding" {
  for_each = {
    for p in local.per_namespace_policies :
    format("%s-%s-%s", replace(replace(p.group, "/@.+$/", ""), ".", "-"), p.role_name, p.namespace) => p
  }

  metadata {
    name      = each.key
    namespace = each.value.namespace
  }

  subject {
    kind      = each.value.kind
    name      = each.value.group
    namespace = each.value.namespace
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = each.value.role_kind
    name      = each.value.role_name
  }

  depends_on = [
    kubernetes_namespace.namespaces,
  ]
}

resource "kubernetes_cluster_role_binding" "cluster_role_binding" {
  for_each = {
    for p in local.cluster_wide_policies :
    format("%s-%s", replace(replace(p.group, "/@.+$/", ""), ".", "-"), p.role_name) => p
  }

  metadata {
    name = each.key
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = each.value.role_kind
    name      = each.value.role_name
  }

  subject {
    kind = each.value.kind
    name = each.value.group
  }
}

resource "kubernetes_cluster_role" "custom_cluster_roles" {
  for_each = {
    for r in var.kubernetes_cluster_roles :
    r.name => r
  }

  metadata {
    name = each.key
  }

  dynamic "rule" {
    for_each = each.value.rules

    content {
      api_groups = rule.value["api_groups"]
      resources  = rule.value["resources"]
      verbs      = rule.value["verbs"]
    }
  }
}

resource "google_monitoring_alert_policy" "istio_nr_host" {
  count = var.gke.enabled && var.gke_enable_platforms_monitoring ? 1 : 0

  project      = var.project_id
  display_name = "[${module.gke[0].name}] Istio proxy got NR (No Route to host) response flags, likely Gateway or Virtualservice are misconfigured"
  combiner     = "OR"

  conditions {
    display_name = "Istio proxy got NR (no route to host) response flags"
    condition_matched_log {
      filter = <<-EOT
        resource.type="k8s_container"
        resource.labels.cluster_name="${module.gke[0].name}"
        resource.labels.namespace_name="istio-system"
        resource.labels.container_name="istio-proxy"
        jsonPayload.response_code="404"
        jsonPayload.response_flags="NR"
        jsonPayload.path=~"^\/manabie\..+|.+\.v\d\..+"
        jsonPayload.method="OPTIONS"
      EOT
    }
  }

  alert_strategy {
    notification_rate_limit {
      period = "300s"
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

data "google_monitoring_notification_channel" "slack_rls" {
  project      = var.project_id
  type         = "slack"
  display_name = "#squad-architecture"
}

resource "google_monitoring_alert_policy" "scan_rls" {
  count = var.gke.enabled && var.gke_enable_platforms_monitoring ? 1 : 0

  project      = var.project_id
  display_name = "[${module.gke[0].name}] Detect failed RLS policy. Check rls-scan container logs"
  combiner     = "OR"

  conditions {
    display_name = "RLS policy failed"
    condition_matched_log {
      filter = <<-EOT
        resource.type="k8s_container"
        resource.labels.cluster_name="${module.gke[0].name}"
        resource.labels.container_name=~".+-scan-rls"
        jsonPayload.msg="rls_scan is error"
        severity>=ERROR
      EOT
    }
  }

  alert_strategy {
    notification_rate_limit {
      period = "300s"
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack_rls.name]
}

resource "google_monitoring_alert_policy" "gke_autoscaler_blocked_scale_down" {
  count = var.gke.enabled ? 1 : 0

  project      = var.project_id
  display_name = "[${module.gke[0].name}] Scale down blocked by pod"
  combiner     = "OR"

  conditions {
    display_name = "Scale down blocked by pod"
    condition_matched_log {
      filter = <<-EOT
        resource.type="k8s_cluster"
        resource.labels.cluster_name="${module.gke[0].name}"
        logName="projects/student-coach-e1e95/logs/container.googleapis.com%2Fcluster-autoscaler-visibility"
        jsonPayload.noDecisionStatus.noScaleDown:*
      EOT
    }
  }

  alert_strategy {
    notification_rate_limit {
      period = "300s"
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

resource "google_monitoring_alert_policy" "gke_node_pool" {
  count = var.gke.enabled && var.gke_enable_resources_monitoring ? 1 : 0

  project      = var.project_id
  display_name = "[${module.gke[0].name}] GKE resource utilization is high"
  combiner     = "OR"

  conditions {
    display_name = "CPU utilization is high"

    condition_threshold {
      filter          = "resource.type = \"k8s_node\" AND resource.labels.cluster_name = \"${module.gke[0].name}\" AND metric.type = \"kubernetes.io/node/cpu/allocatable_utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.9

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }

      trigger {
        count = 1
      }
    }
  }

  conditions {
    display_name = "Memory utilization is high"

    condition_threshold {
      filter          = "resource.type = \"k8s_node\" AND resource.labels.cluster_name = \"${module.gke[0].name}\" AND metric.type = \"kubernetes.io/node/memory/allocatable_utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.9

      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_MEAN"
      }

      trigger {
        count = 1
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

resource "google_monitoring_alert_policy" "prometheus_alertmanager" {
  count = var.gke.enabled && var.gke_enable_platforms_monitoring ? 1 : 0

  project      = var.project_id
  display_name = "[${module.gke[0].name}] Prometheus or Alertmanager pods in GKE are crash looping"
  combiner     = "OR"

  conditions {
    display_name = "Prometheus server is crash looping"

    condition_threshold {
      filter = <<-EOT
        metric.type="kubernetes.io/container/restart_count"
        resource.type="k8s_container"
        resource.label."container_name"="prometheus-server"
        resource.label."cluster_name"="${module.gke[0].name}"
      EOT

      duration        = "0s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0

      aggregations {
        alignment_period   = "600s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  conditions {
    display_name = "Alertmanager is crash looping"

    condition_threshold {
      filter = <<-EOT
        metric.type="kubernetes.io/container/restart_count"
        resource.type="k8s_container"
        resource.label."container_name"="prometheus-alertmanager"
        resource.label."cluster_name"="${module.gke[0].name}"
      EOT

      duration        = "0s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0

      aggregations {
        alignment_period   = "600s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }

  conditions {
    display_name = "Prometheus server keeps restarting"

    condition_threshold {
      filter = <<-EOT
        metric.type="kubernetes.io/container/uptime"
        resource.type="k8s_container"
        resource.label."container_name"="prometheus-server"
        resource.label."cluster_name"="${module.gke[0].name}"
      EOT

      duration        = "600s"
      comparison      = "COMPARISON_LT"
      threshold_value = 600 // 10 minutes

      aggregations {
        per_series_aligner   = "ALIGN_MAX"
        alignment_period     = "300s"
        group_by_fields      = ["resource.label.container_name"]
        cross_series_reducer = "REDUCE_MAX"
      }
    }
  }

  notification_channels = [data.google_monitoring_notification_channel.slack.name]
}

data "google_monitoring_notification_channel" "slack_user_mgmt" {
  project      = var.project_id
  type         = "slack"
  display_name = "Camp User alerts"
}
