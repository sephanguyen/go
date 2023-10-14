locals {
  roles = distinct(flatten([
    for proj_roles in var.service_accounts[*].roles : [
      for proj, roles in proj_roles : compact(roles)
    ]
  ]))

  bucket_roles = distinct(flatten([
    for bucket_roles in var.service_accounts[*].bucket_roles : [
      for bucket, roles in bucket_roles : compact(roles)
    ]
    if bucket_roles != null
  ]))

  identity_namespaces = distinct(flatten(var.service_accounts[*].identity_namespaces))

  service_accounts = [
    for sa in var.service_accounts : {
      project = sa.project
      name    = sa.name
    }
    if sa.name != null
  ]

  service_accounts_roles = flatten([
    for role in local.roles : [
      for sa in var.service_accounts : [
        for project, roles in sa.roles : {
          project       = sa.project
          name          = sa.name
          grant_project = project
          role          = role
        }
        if contains(roles, role)
      ]
    ]
  ])

  service_accounts_bucket_roles = flatten([
    for role in coalesce(local.bucket_roles, []) : [
      for sa in var.service_accounts : [
        for bucket, roles in sa.bucket_roles : {
          project      = sa.project
          name         = sa.name
          grant_bucket = bucket
          role         = role
        }
        if contains(roles, role)
      ]
      if sa.bucket_roles != null
    ]
  ])

  workload_identities = flatten([
    for ns in local.identity_namespaces : [
      for sa in var.service_accounts : {
        project   = sa.project
        name      = sa.name
        namespace = ns
      }
      if ns != "" && contains(sa.identity_namespaces, ns)
    ]
  ])

  service_accounts_impersonation = flatten([
    for sa in var.service_accounts : [
      for i in coalesce(sa.impersonations, []) : {
        principal_sa          = sa.name
        principal_project     = sa.project
        impersonation_sa      = i.name
        impersonation_project = i.project
        impersonation_role    = i.role
      }
    ]
  ])
}

resource "google_service_account" "service_accounts" {
  for_each = {
    for sa in local.service_accounts :
    "${sa.project}.${sa.name}" => sa
  }

  project    = each.value.project
  account_id = lower(each.value.name)
}

resource "google_project_iam_member" "project" {
  for_each = {
    for sa in local.service_accounts_roles :
    "${sa.grant_project}.${sa.name}.${sa.role}" => sa
  }

  project = each.value.grant_project
  role    = each.value.role
  member  = "serviceAccount:${lookup(google_service_account.service_accounts, "${each.value.project}.${each.value.name}", {}).email}"
}

resource "google_service_account_iam_member" "workload_identities" {
  for_each = {
    for w in local.workload_identities :
    "${w.project}.${w.namespace}.${w.name}" => w
  }

  service_account_id = format(
    "projects/%s/serviceAccounts/%s",
    each.value.project,
    lookup(google_service_account.service_accounts, "${each.value.project}.${each.value.name}", {}).email,
  )
  role   = "roles/iam.workloadIdentityUser"
  member = format("serviceAccount:%s[%s/%s]", var.gke_identity_namespace, each.value.namespace, each.value.name)
}

resource "google_service_account_iam_member" "service_account_impersonation" {
  for_each = {
    for s in local.service_accounts_impersonation :
    "${s.principal_project}.${s.principal_sa}.${s.impersonation_project}.${s.impersonation_sa}" => s
  }

  service_account_id = format(
    "projects/%s/serviceAccounts/%s",
    lookup(
      google_service_account.service_accounts,
      "${each.value.impersonation_project}.${each.value.impersonation_sa}",
      {},
    ).project,
    lookup(
      google_service_account.service_accounts,
      "${each.value.impersonation_project}.${each.value.impersonation_sa}",
      {},
    ).email,
  )
  role = each.value.impersonation_role
  member = format(
    "serviceAccount:%s",
    lookup(
      google_service_account.service_accounts,
      "${each.value.principal_project}.${each.value.principal_sa}",
      {},
    ).email,
  )
}

resource "google_storage_bucket_iam_member" "bucket" {
  for_each = {
    for sa in local.service_accounts_bucket_roles :
    "${sa.grant_bucket}.${sa.name}.${sa.role}" => sa
  }

  bucket = each.value.grant_bucket
  role   = each.value.role
  member = "serviceAccount:${lookup(google_service_account.service_accounts, "${each.value.project}.${each.value.name}", {}).email}"
}

locals {
  # get service account that has role Firebase Admin and generate HMAC key
  # for that service account, since Firebase Admin role also has enough
  # permissions on Cloud Storage resources.
  # See https://cloud.google.com/iam/docs/understanding-roles#firebase-roles.
  storage_sa = var.create_storage_hmac_key ? [
    for sa in local.service_accounts_roles : {
      project = sa.project
      name    = sa.name
    }
    if sa.role == "roles/firebase.admin"
  ] : []
}

resource "google_storage_hmac_key" "storage" {
  count = length(local.storage_sa) > 0 ? 1 : 0

  service_account_email = lookup(
    google_service_account.service_accounts,
    "${local.storage_sa[0].project}.${local.storage_sa[0].name}",
    {},
  ).email
}

resource "google_service_account" "cloudconvert" {
  count = var.cloudconvert != null ? 1 : 0

  account_id = var.cloudconvert.service_account
}

resource "google_storage_bucket_iam_member" "cloudconvert" {
  count = var.cloudconvert != null ? 1 : 0

  bucket = var.cloudconvert.bucket
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${google_service_account.cloudconvert[0].email}"
}

# cloudconvert service requires the service account JSON credentials
# so it can upload the converted images to the Cloud Storage
resource "google_service_account_key" "cloudconvert" {
  count = var.cloudconvert != null ? 1 : 0

  service_account_id = google_service_account.cloudconvert[0].name
}
