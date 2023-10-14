locals {
  namespaces = var.gke_rbac.enabled ? compact(flatten(var.gke_rbac.policies[*].namespaces)) : []

  policies = var.gke_rbac.enabled ? flatten([
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

  rbac_policies = var.gke_rbac.enabled ? distinct(flatten([
    for level, values in try(var.rbac_roles.policies, {}) : [
      for _, roles in values : [
        for role in roles : [
          for ns in role.namespaces : {
            kind      = role.kind
            group     = role.group
            namespace = ns
            role_kind = role.role_kind
            role_name = role.role_name
            rules     = role.rules
          }
        ]
      ]
    ]
  ])) : []

  custom_roles = [
    for policy in local.rbac_policies : policy
    if policy.role_kind == "Role"
  ]
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

resource "kubernetes_role" "roles" {
  for_each = {
    for r in local.custom_roles :
    format("%s-%s-%s", replace(replace(r.group, "/@.+$/", ""), ".", "-"), r.role_name, r.namespace) => r
  }

  metadata {
    name      = each.key
    namespace = each.value.namespace
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

resource "kubernetes_role_binding" "role_binding" {
  for_each = {
    for p in local.policies :
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

// Don't need to use below block if we merge with current policies. Just use above one
resource "kubernetes_role_binding" "rbac_role_binding" {
  for_each = {
    for p in local.rbac_policies :
    format("%s-%s-%s-%s", p.role_kind, p.role_name, p.namespace, replace(replace(p.group, "/@.+$/", ""), ".", "-")) => p
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
    kubernetes_role.roles,
  ]
}
