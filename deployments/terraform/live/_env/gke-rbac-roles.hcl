locals {
  # Define RBAC policies based on access level.
  #
  # Note that namespaces defined here is not the "real" namespace name,
  # they're just placeholder values, and will be replaced in the corresponding
  # file {stag,uat,prod}-apps.hcl. That's because namespace name depend on
  # current organization and environment, but we don't have those values here.
  rbac_policies = {
    low = {
      stag = [
        {
          kind      = "Group"
          group     = "tech-func-platform-low@manabie.com"
          role_kind = "ClusterRole"
          role_name = "view"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      uat = [
        {
          kind      = "Group"
          group     = "tech-func-platform-low@manabie.com"
          role_kind = "ClusterRole"
          role_name = "view"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      prod = [
        {
          kind      = "Group"
          group     = "tech-func-platform-low@manabie.com"
          role_kind = "Role"
          role_name = "platform-low"
          namespaces = [
            "services",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs"]
              verbs      = ["get", "list", "watch"]
            }
          ]
        },
        {
          kind      = "Group"
          group     = "tech-func-platform-low@manabie.com"
          role_kind = "Role"
          role_name = "platform-low"
          namespaces = [
            "machine-learning",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs"]
              verbs      = ["get", "list", "watch"]
            }
          ]
        }
      ],
    }
    moderate = {
      stag = [
        {
          kind      = "Group"
          group     = "tech-func-platform-moderate@manabie.com"
          role_kind = "ClusterRole"
          role_name = "view"
          namespaces = [
            "services",
          ]
        },
        {
          kind      = "Group"
          group     = "tech-func-platform-moderate@manabie.com"
          role_kind = "ClusterRole"
          role_name = "edit"
          namespaces = [
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      uat = [
        {
          kind      = "Group"
          group     = "tech-func-platform-moderate@manabie.com"
          role_kind = "ClusterRole"
          role_name = "view"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      prod = [
        {
          kind      = "Group"
          group     = "tech-func-platform-moderate@manabie.com"
          role_kind = "Role"
          role_name = "platform-moderate"
          namespaces = [
            "elastic",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs"]
              verbs      = ["get", "list", "watch", "create", "update", "patch"]
            }
          ]
        },
        {
          kind      = "Group"
          group     = "tech-func-platform-moderate@manabie.com"
          role_kind = "Role"
          role_name = "platform-moderate"
          namespaces = [
            "machine-learning",
            "elastic",
            "kafka",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs"]
              verbs      = ["get", "list", "watch", "create", "update", "patch"]
            }
          ]
        }
      ]
    }
    high = {
      stag = [
        {
          kind      = "Group"
          group     = "tech-func-platform-high@manabie.com"
          role_kind = "ClusterRole"
          role_name = "admin"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      uat = [
        {
          kind      = "Group"
          group     = "tech-func-platform-high@manabie.com"
          role_kind = "ClusterRole"
          role_name = "admin"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      prod = [
        {
          kind      = "Group"
          group     = "tech-func-platform-high@manabie.com"
          role_kind = "Role"
          role_name = "platform-high"
          namespaces = [
            "services",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs", "secrets"]
              verbs      = ["get", "list", "watch", "create"]
            }
          ]
        },
        {
          kind      = "Group"
          group     = "tech-func-platform-high@manabie.com"
          role_kind = "Role"
          role_name = "platform-high"
          namespaces = [
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
          rules = [
            {
              api_groups = [""]
              resources  = ["nodes", "pods", "deployments", "configmaps", "endpoints", "services", "daemonsets", "statefulsets", "cronjobs", "jobs", "secrets"]
              verbs      = ["get", "list", "watch", "create", "update", "patch"]
            }
          ]
        }
      ]
    }
    super = {
      stag = [
        {
          kind      = "Group"
          group     = "tech-func-platform-super@manabie.com"
          role_kind = "ClusterRole"
          role_name = "cluster-admin"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      uat = [
        {
          kind      = "Group"
          group     = "tech-func-platform-super@manabie.com"
          role_kind = "ClusterRole"
          role_name = "cluster-admin"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
      prod = [
        {
          kind      = "Group"
          group     = "tech-func-platform-super@manabie.com"
          role_kind = "ClusterRole"
          role_name = "cluster-admin"
          namespaces = [
            "services",
            "nats-jetstream",
            "elastic",
            "kafka",
          ]
        }
      ]
    }
  }
}
