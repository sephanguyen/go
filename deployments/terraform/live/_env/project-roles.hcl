locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  // each GCP project will have N custom roles, where N is calculated by:
  //    total functions * total access levels
  //
  //  where functions are:
  //    - data
  //    - platform
  //    - backend
  //    - frontend
  //  and access levels:
  //    - low
  //    - moderate
  //    - high
  //    - super
  //
  //  so we will have these custom roles:
  //
  //  - customroles.data.low
  //  - customroles.data.moderate
  //  - customroles.data.high
  //  - customroles.data.super
  //
  //  - customroles.platform.low
  //  - customroles.platform.moderate
  //  ...
  //  ...
  //  ...
  //  and so on.

  custom_perms = distinct([
    // These permissions come from Cloud SQL Viewer role
    // exclude resourcemanager.projects.list permission,
    // because we can't create custom role with that permission.
    // See https://cloud.google.com/iam/docs/understanding-custom-roles#known_limitations.
    "cloudsql.backupRuns.get",
    "cloudsql.backupRuns.list",
    "cloudsql.databases.get",
    "cloudsql.databases.list",
    "cloudsql.instances.export",
    "cloudsql.instances.get",
    "cloudsql.instances.list",
    "cloudsql.instances.listServerCas",
    "cloudsql.instances.listTagBindings",
    "cloudsql.sslCerts.get",
    "cloudsql.sslCerts.list",
    "cloudsql.users.list",
    "recommender.cloudsqlIdleInstanceRecommendations.get",
    "recommender.cloudsqlIdleInstanceRecommendations.list",
    "recommender.cloudsqlInstanceActivityInsights.get",
    "recommender.cloudsqlInstanceActivityInsights.list",
    "recommender.cloudsqlInstanceCpuUsageInsights.get",
    "recommender.cloudsqlInstanceCpuUsageInsights.list",
    "recommender.cloudsqlInstanceDiskUsageTrendInsights.get",
    "recommender.cloudsqlInstanceDiskUsageTrendInsights.list",
    "recommender.cloudsqlInstanceMemoryUsageInsights.get",
    "recommender.cloudsqlInstanceMemoryUsageInsights.list",
    "recommender.cloudsqlInstanceOutOfDiskRecommendations.get",
    "recommender.cloudsqlInstanceOutOfDiskRecommendations.list",
    "recommender.cloudsqlOverprovisionedInstanceRecommendations.get",
    "recommender.cloudsqlOverprovisionedInstanceRecommendations.list",
    "resourcemanager.projects.get",
    "serviceusage.quotas.get",
    "serviceusage.services.get",
    "serviceusage.services.list",

    // These permissions come from Firebase Analytics Viewer role
    // exclude resourcemanager.projects.list permission,
    // because we can't create custom role with that permission.
    // See https://cloud.google.com/iam/docs/understanding-custom-roles#known_limitations.
    "cloudnotifications.activities.list",
    "firebase.billingPlans.get",
    "firebase.clients.get",
    "firebase.clients.list",
    "firebase.links.list",
    "firebase.playLinks.get",
    "firebase.playLinks.list",
    "firebase.projects.get",
    "firebaseanalytics.resources.googleAnalyticsReadAndAnalyze",
    "firebaseextensions.configs.list",
    "resourcemanager.projects.get",
    "resourcemanager.projects.getIamPolicy",

    // These permissions come from Monitoring Viewer role
    // exclude resourcemanager.projects.list permission,
    // because we can't create custom role with that permission.
    // See https://cloud.google.com/iam/docs/understanding-custom-roles#known_limitations.
    "cloudnotifications.activities.list",
    "monitoring.alertPolicies.get",
    "monitoring.alertPolicies.list",
    "monitoring.dashboards.get",
    "monitoring.dashboards.list",
    "monitoring.groups.get",
    "monitoring.groups.list",
    "monitoring.metricDescriptors.get",
    "monitoring.metricDescriptors.list",
    "monitoring.monitoredResourceDescriptors.get",
    "monitoring.monitoredResourceDescriptors.list",
    "monitoring.notificationChannelDescriptors.get",
    "monitoring.notificationChannelDescriptors.list",
    "monitoring.notificationChannels.get",
    "monitoring.notificationChannels.list",
    "monitoring.publicWidgets.get",
    "monitoring.publicWidgets.list",
    "monitoring.services.get",
    "monitoring.services.list",
    "monitoring.slos.get",
    "monitoring.slos.list",
    "monitoring.timeSeries.list",
    "monitoring.uptimeCheckConfigs.get",
    "monitoring.uptimeCheckConfigs.list",
    "opsconfigmonitoring.resourceMetadata.list",
    "resourcemanager.projects.get",
    "stackdriver.projects.get",

    // These permissions come from Cloud Build Viewer roles (roles/cloudbuild.builds.viewer)
    // excluding resourcemanager.projects.list permissions,
    // because we can't create custom role with that permission.
    "cloudbuild.builds.get",
    "cloudbuild.builds.list",
    "remotebuildexecution.blobs.get",
    "resourcemanager.projects.get",

    // These permissions come from Cloud Trace User roles (roles/cloudtrace.user)
    // excluding resourcemanager.projects.list permissions,
    // because we can't create custom role with that permission.
    "cloudtrace.insights.get",
    "cloudtrace.insights.list",
    "cloudtrace.stats.get",
    "cloudtrace.tasks.create",
    "cloudtrace.tasks.delete",
    "cloudtrace.tasks.get",
    "cloudtrace.tasks.list",
    "cloudtrace.traces.get",
    "cloudtrace.traces.list",

    // These permissions come from Kubernetes Engine Cluster Viewer roles (roles/container.clusterViewer)
    // excluding resourcemanager.projects.list permissions,
    // because we can't create custom role with that permission.
    "container.clusters.get",
    "container.clusters.list",
  ])

  # These permissions come from BigQuery User role
  # exclude resourcemanager.projects.list permission,
  # because we can't create custom role with that permission.
  # See https://cloud.google.com/iam/docs/understanding-custom-roles#known_limitations.
  bigquery_user_perms = [
    "bigquery.bireservations.get",
    "bigquery.capacityCommitments.get",
    "bigquery.capacityCommitments.list",
    "bigquery.config.get",
    "bigquery.datasets.create",
    "bigquery.datasets.get",
    "bigquery.datasets.getIamPolicy",
    "bigquery.jobs.create",
    "bigquery.jobs.list",
    "bigquery.models.list",
    "bigquery.readsessions.create",
    "bigquery.readsessions.getData",
    "bigquery.readsessions.update",
    "bigquery.reservationAssignments.list",
    "bigquery.reservationAssignments.search",
    "bigquery.reservations.get",
    "bigquery.reservations.list",
    "bigquery.routines.list",
    "bigquery.savedqueries.get",
    "bigquery.savedqueries.list",
    "bigquery.tables.list",
    "bigquery.transfers.get",
    "bigquerymigration.translation.translate",
    "resourcemanager.projects.get",
  ]

  backend_custom_perms = distinct(concat(
    local.custom_perms,
    [
      // These permissions come from Cloud Profiler User roles (roles/cloudprofiler.user)
      // excluding resourcemanager.projects.list permissions,
      // because we can't create custom role with that permission.
      "cloudprofiler.profiles.list",
      "resourcemanager.projects.get",
      "serviceusage.quotas.get",
      "serviceusage.services.get",
      "serviceusage.services.list",
    ],
  ))

  platform_custom_perms = distinct(concat(
    local.custom_perms,
    [
      // View service accounts permissions.
      // See https://console.cloud.google.com/iam-admin/roles/details/roles%3Ciam.serviceAccountViewer
      "iam.serviceAccountKeys.get",
      "iam.serviceAccountKeys.list",
      "iam.serviceAccounts.get",
      "iam.serviceAccounts.getIamPolicy",
      "iam.serviceAccounts.list",
      "resourcemanager.projects.get",

      // These permissions come from Cloud Profiler User roles (roles/cloudprofiler.user)
      // excluding resourcemanager.projects.list permissions,
      // because we can't create custom role with that permission.
      "cloudprofiler.profiles.list",
      "resourcemanager.projects.get",
      "serviceusage.quotas.get",
      "serviceusage.services.get",
      "serviceusage.services.list",
    ]
  ))

  roles = [
    // DATA roles
    {
      id          = "customroles.data.low"
      title       = "Data Role for Low Access Level"
      description = ""
      base_roles  = []
      permissions = distinct(concat(
        local.custom_perms,
        local.bigquery_user_perms,
      ))
    },
    {
      id          = "customroles.data.moderate"
      title       = "Data Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
      ]
      permissions = distinct(concat(
        local.custom_perms,
        local.bigquery_user_perms,
      ))
    },
    {
      id          = "customroles.data.high",
      title       = "Data Role for High Access Level",
      description = "",
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
      ]
      permissions = distinct(concat(
        local.custom_perms,
        local.bigquery_user_perms,
      ))
    },
    {
      id          = "customroles.data.super"
      title       = "Data Role for Super Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
      ]
      permissions = distinct(concat(
        local.custom_perms,
        local.bigquery_user_perms,
        [
          // Firebase Analytics Admin permissions.
          // See https://console.cloud.google.com/iam-admin/roles/details/roles%3Cfirebase.analyticsAdmin
          "firebaseanalytics.resources.googleAnalyticsEdit",
          "firebaseanalytics.resources.googleAnalyticsReadAndAnalyze",
          "firebaseextensions.configs.list",
        ],
      ))
    },

    // BACKEND roles
    {
      id          = "customroles.backend.low"
      title       = "Backend Role for Low Access Level"
      description = ""
      base_roles  = []
      permissions = local.custom_perms
    },
    {
      id          = "customroles.backend.moderate"
      title       = "Backend Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.backend_custom_perms
    },
    {
      id          = "customroles.backend.high",
      title       = "Backend Role for High Access Level",
      description = "",
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
        "roles/logging.privateLogViewer",
      ]
      permissions = local.backend_custom_perms
    },
    {
      id          = "customroles.backend.super"
      title       = "Backend Role for Super Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
        "roles/logging.privateLogViewer",
      ]
      permissions = local.backend_custom_perms
    },

    // PLATFORM roles
    {
      id          = "customroles.platform.low"
      title       = "Platform Role for Low Access Level"
      description = ""
      base_roles  = []
      permissions = local.platform_custom_perms
    },
    {
      id          = "customroles.platform.moderate"
      title       = "Platform Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.platform_custom_perms
    },
    {
      id          = "customroles.platform.high",
      title       = "Platform Role for High Access Level",
      description = "",
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
        "roles/logging.privateLogViewer",
      ]
      permissions = distinct(concat(
        local.platform_custom_perms,
        [
          // Working with Cloud Build; the following permissions are taken from:
          // - roles/cloudbuild.editor
          // - roles/cloudbuild.approver
          // See https://cloud.google.com/build/docs/iam-roles-permissions#predefined_roles
          "cloudbuild.builds.get",
          "cloudbuild.builds.list",
          "cloudbuild.builds.create",
          "cloudbuild.builds.update",
          "cloudbuild.builds.approve",

          // Working with Workload Identity Federation;
          // Allowing user go to this side: https://console.cloud.google.com/iam-admin/workload-identity-pools
          "iam.workloadIdentityPoolProviders.list",
          "iam.workloadIdentityPools.list",
        ],
      ))
    },
    {
      id          = "customroles.platform.super"
      title       = "Platform Role for Super Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
        "roles/logging.privateLogViewer",
      ]
      permissions = distinct(concat(
        local.platform_custom_perms,
        [
          // Working with Cloud Build; the following permissions are taken from:
          // - roles/cloudbuild.editor
          // - roles/cloudbuild.approver
          // See https://cloud.google.com/build/docs/iam-roles-permissions#predefined_roles
          "cloudbuild.builds.get",
          "cloudbuild.builds.list",
          "cloudbuild.builds.create",
          "cloudbuild.builds.update",
          "cloudbuild.builds.approve",

          // To set permissions in IAM & Admin
          "resourcemanager.projects.setIamPolicy",

          // Working with Workload Identity Federation;
          // Allowing user go to this side: https://console.cloud.google.com/iam-admin/workload-identity-pools
          "iam.workloadIdentityPoolProviders.list",
          "iam.workloadIdentityPools.list",
        ],
      ))
    },

    // MOBILE roles
    {
      id          = "customroles.mobile.low"
      title       = "Mobile Role for Low Access Level"
      description = ""
      base_roles  = []
      permissions = local.custom_perms
    },
    {
      id          = "customroles.mobile.moderate"
      title       = "Mobile Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },
    {
      id          = "customroles.mobile.high",
      title       = "Mobile Role for High Access Level",
      description = "",
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },
    {
      id          = "customroles.mobile.super"
      title       = "Mobile Role for Super Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },

    // WEB roles
    {
      id          = "customroles.web.low"
      title       = "Web Role for Low Access Level"
      description = ""
      base_roles  = []
      permissions = local.custom_perms
    },
    {
      id          = "customroles.web.moderate"
      title       = "Web Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },
    {
      id          = "customroles.web.high",
      title       = "Web Role for High Access Level",
      description = "",
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },
    {
      id          = "customroles.web.super"
      title       = "Web Role for Super Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
        "roles/logging.viewer",
      ]
      permissions = local.custom_perms
    },
    {
      id          = "customroles.cloudbuild.approval"
      title       = "Cloud build approval permission"
      description = "Grant cloud build approval permission for tech-lead only"
      base_roles  = []
      permissions = [
        "cloudbuild.builds.approve",
      ]
    }
  ]
}

inputs = {
  project_id = local.env_vars.locals.project_id
  roles      = local.roles
}
