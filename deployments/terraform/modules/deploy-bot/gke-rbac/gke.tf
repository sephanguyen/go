
# Some namespaces, such as istio-system, are shared
locals {
  shared_namespaces = var.configure_common_namespaces ? [
    "istio-system",
  ] : []
}

resource "kubernetes_role_binding" "dorp_deploy_bot_role_bindings_common_namespaces" {
  for_each = toset(local.shared_namespaces)

  metadata {
    name      = "dorp-deploy-bot-cluster-admin-${each.value}"
    namespace = each.value
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }

  subject {
    name      = var.dorp_deploy_bot_service_account.email
    kind      = "User"
    namespace = each.value
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "kubernetes_role_binding" "dorp_deploy_bot_role_bindings" {
  for_each = toset(local.preproduction_namespaces)

  metadata {
    name      = "dorp-deploy-bot-cluster-admin-${each.value}"
    namespace = each.value
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }

  subject {
    name      = var.dorp_deploy_bot_service_account.email
    kind      = "User"
    namespace = each.value
    api_group = "rbac.authorization.k8s.io"
  }

  depends_on = [
    kubernetes_namespace.dorp_namespaces
  ]
}

resource "kubernetes_role_binding" "platform_func_role_bindings" {
  for_each = toset(local.preproduction_namespaces)

  metadata {
    name      = "tech-func-platform-cluster-admin-${each.value}"
    namespace = each.value
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }

  subject {
    name      = "tech-func-platform@manabie.com"
    kind      = "Group"
    namespace = each.value
    api_group = "rbac.authorization.k8s.io"
  }

  depends_on = [
    kubernetes_namespace.dorp_namespaces
  ]
}
