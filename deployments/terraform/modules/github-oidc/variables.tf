variable "project_id" {
  type        = string
  description = "GCP project to create Workload Identity Pool"
}

variable "pool_id" {
  type        = string
  description = "Workload Identity Pool ID"
}

variable "deploy_bot" {
  type = object({
    iam = optional(object({
      name       = string
      project_id = string
      roles      = list(string)
    }))
    service_account_id = string
    wif = object({
      provider_id           = optional(string, "gh-action-provider")
      provider_display_name = optional(string, "Deploy Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Deploy Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration of a Google Cloud Workload Identity Federation provider.
This provider is then used for a Github Workflow to deploy to Kubernetes.
EOF
}

variable "build_bot" {
  type = object({
    iam = object({
      name       = string
      project_id = string
      roles      = map(list(string))
    })
    wif = object({
      provider_id           = optional(string, "build-bot-provider")
      provider_display_name = optional(string, "Build Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Build Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration for a bot account which is mainly used to build and push Docker images
to GCP container/artifact registries.
Argument reference:
- iam: describes the service account of this bot
  - name: service account name
  - project_id: location to create the service account
  - roles: IAM roles to grant for the bot's service account
- wif: describes the workload identity federation that this bot will use
  - project_id: location to get the WIF pool/create the provider
  - attribute_conditions: condition for the provider
  - attribute_mapping: mapping for the provider, rarely changed
  - allowed_audiences: as name suggests
  - github_repository: as name suggests, should be the repository where the Github workflows are activated
EOF
}

variable "integration_test_bot" {
  type = object({
    iam = object({
      name       = string
      project_id = string
      roles      = map(list(string))
    })
    wif = object({
      provider_id           = optional(string, "integration-test-bot-provider")
      provider_display_name = optional(string, "Integration Test Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Integration Test Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration for a bot account which is mainly used to pull Docker images
when running integration test on backend's Github Action.
The input arguments are identical to those of `build_bot`.
EOF
}

variable "unleash_decryptor_bot" {
  type = object({
    iam = object({
      name       = string
      project_id = string
      roles      = map(list(string))
    })
    wif = object({
      provider_id           = optional(string, "unleash-decryptor-bot")
      provider_display_name = optional(string, "Unleash Decryptor Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Unleash Decryptor Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration for a bot account used to decryptor secrets to retrieve
the admin token required to work with Unleash on production.
The input arguments are identical to those of `build_bot`.
EOF
}

variable "dorp_deploy_bot" {
  type = object({
    iam                = optional(object({})) # unused
    service_account_id = string
    wif = object({
      provider_id           = optional(string, "dorp-deploy-bot-provider")
      provider_display_name = optional(string, "Dorp Deploy Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Dorp Deploy Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration of a Google Cloud Workload Identity Federation provider.
This provider is then used for a Github Workflow to deploy to Kubernetes.
EOF
}


variable "mfe_deploy_bot" {
  type = object({
    iam = object({
      name         = string
      project_id   = string
      roles        = map(list(string))
      bucket_names = list(string)
    })
    wif = object({
      provider_id           = optional(string, "mfe-deploy-provider")
      provider_display_name = optional(string, "MFE Deploy Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for MFE Deploy Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"       = "assertion.sub"
        "attribute.actor"      = "assertion.actor"
        "attribute.repository" = "assertion.repository"
        "attribute.workflow"   = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
      github_repository = optional(string, "manabie-com/backend")
    })
  })

  default     = null
  description = <<EOF
Configuration of a Google Cloud Workload Identity Federation provider.
This provider is then used for a Github Workflow to deploy to Kubernetes.
EOF
}


variable "get_release_tag_bot" {
  type = object({
    iam = optional(object({
      name       = string
      project_id = string
      roles      = list(string)
    }))
    service_account_id = string
    wif = object({
      provider_id           = optional(string, "get-release-tag-provider")
      provider_display_name = optional(string, "Get Release Tag Bot Provider")
      provider_description  = optional(string, "Workload Identity Federation Provider for Get Release Tag Bot. Managed by Terraform.")
      attribute_condition   = optional(string)
      attribute_mapping = optional(map(any), {
        "google.subject"     = "assertion.sub"
        "attribute.actor"    = "assertion.actor"
        "attribute.workflow" = "assertion.workflow"
      })
      allowed_audiences = optional(list(string), [])
      issuer_uri        = optional(string, "https://token.actions.githubusercontent.com")
    })
  })

  default     = null
  description = <<EOF
Configuration of a Google Cloud Workload Identity Federation provider.
This provider is then used for a Github Workflow to deploy to Kubernetes.
EOF
}