locals {
  preproduction_namespaces = var.configure_dorp_namespaces ? [
    "dorp-${var.org}-elastic",
    "dorp-${var.org}-nats-jetstream",
    "dorp-${var.org}-kafka",
    "dorp-${var.org}-services",
  ] : []
}

resource "kubernetes_namespace" "dorp_namespaces" {
  for_each = toset(local.preproduction_namespaces)

  metadata {
    name = each.value
  }

  lifecycle {
    ignore_changes = [
      metadata,
    ]
  }
}
