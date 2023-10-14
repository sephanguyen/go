# Creates google groups and memberships for squads and functions
# https://groups.google.com/my-groups
#
# Also gives squad managers kms powers for their owned services
module "group_dev" {
  source  = "terraform-google-modules/group/google"
  version = "~> 0.1"

  customer_id  = "C00ziiz00"
  id           = "dev@manabie.com"
  display_name = "dev"
  description  = "dev managed by terraform script"
  owners       = distinct(compact([for mem in var.members : mem.email if !mem.google_group.disabled && mem.github.role == "admin"]))
  members      = distinct(compact([for mem in var.members : mem.email if !mem.google_group.disabled && mem.github.role == "member"]))
}

locals {
  google_group_function_members = { for function in local.github_functions :
    function => flatten([
      for member in var.members : [
        for func in member.functions : member.email
        if func.name == function && !member.google_group.disabled && func.role == "member"
      ]
    ])
  }

  google_group_function_platform_members = { for level in local.access_levels :
    level => flatten([
      for member in var.members : [
        for func in member.functions : member.email
        if func.name == "platform" && !member.google_group.disabled && func.role == "member" && func.access_level == level
      ]
    ])
  }

  google_group_function_managers = { for function in local.github_functions :
    function => flatten([
      for member in var.members : [
        for func in member.functions : member.email
        if func.name == function && !member.google_group.disabled && func.role == "maintainer"
      ]
      ]
    )
  }

  google_group_function_owners = { for function in local.github_functions :
    function => flatten([
      for member in var.members : [
        for func in member.functions : member.email
        if func.name == function && !member.google_group.disabled && func.role == "manager"
      ]
      ]
    )
  }

  google_group_squad_members = { for squad in local.github_squads :
    squad => flatten([
      for member in var.members : [
        for func in member.squads : member.email
        if func.name == squad && !member.google_group.disabled && func.role == "member"
      ]
      ]
    )
  }

  google_group_squad_managers = { for squad in local.github_squads :
    squad => flatten([
      for member in var.members : [
        for func in member.squads : member.email
        if func.name == squad && !member.google_group.disabled && func.role == "maintainer"
      ]
      ]
    )
  }

  google_group_squad_owners = { for squad in local.github_squads :
    squad => flatten([
      for member in var.members : [
        for func in member.squads : member.email
        if func.name == squad && !member.google_group.disabled && func.role == "manager"
      ]
      ]
    )
  }
}

module "group_funcs" {
  source   = "terraform-google-modules/group/google"
  version  = "~> 0.1"
  for_each = local.google_group_function_members

  customer_id  = "C00ziiz00"
  id           = "tech-func-${each.key}@manabie.com"
  display_name = "tech-func-${each.key}"
  description  = "tech-func-${each.key} managed by terraform script"
  owners       = local.google_group_function_owners[each.key]
  managers     = local.google_group_function_managers[each.key]
  members      = each.value
}

module "group_funcs_platform" {
  source   = "terraform-google-modules/group/google"
  version  = "~> 0.1"
  for_each = local.google_group_function_platform_members

  customer_id  = "C00ziiz00"
  id           = "tech-func-platform-${each.key}@manabie.com"
  display_name = "tech-func-platform-${each.key}"
  description  = "tech-func-platform-${each.key} managed by terraform script"
  owners       = local.google_group_function_owners["platform"]
  managers     = local.google_group_function_managers["platform"]
  members      = each.value
}

module "group_squads" {
  source   = "terraform-google-modules/group/google"
  version  = "~> 0.1"
  for_each = local.google_group_squad_members

  customer_id  = "C00ziiz00"
  id           = "tech-squad-${each.key}@manabie.com"
  display_name = "tech-squad-${each.key}"
  description  = "tech-squad-${each.key} managed by terraform script"
  owners       = local.google_group_squad_owners[each.key]
  managers     = local.google_group_squad_managers[each.key]
  members      = each.value
}

data "google_cloud_identity_groups" "all_groups" {
  parent = "customers/C00ziiz00"
}

locals {
  squad_list = distinct(flatten([
    for m in var.members : [
      for s in m.squads : s.name
    ]
  ]))

  /**
  [{ squad = "adobo", manager = "manager_email@manabie.com"}, ...]
  */
  squad_managers = concat(flatten([
    for squad_name in local.squad_list : [
      for m in var.members : [
        for s in m.squads : { squad = squad_name, manager = m.email } if s.role == "manager" && s.name == squad_name
      ]
    ]
  ]))

  /**
  [{ squad = "adobo", manager = "manager_email@manabie.com", service = "bob", env = "stag" }, ...]
  */
  ownerships = flatten([
    for sm in local.squad_managers : [
      for o in var.service_ownerships : {
        squad   = sm.squad
        manager = sm.manager
        service = o.service
        env     = o.env
      } if o.squad == sm.squad
    ]
  ])
}

# This resource allows the managers of each squad to encrypt/decrypt secrets
# of their services.
#   - "managers" are specified using var.members[].squads[].role == "manager".
#   - service ownerships are specified using var.service_ownerships.
resource "google_cloud_identity_group_membership" "kms_encrypt_decrypters" {
  for_each = {
    for o in local.ownerships :
    "${o.squad}.${o.manager}.${o.env}.${o.service}" => o
  }

  # This line assumes that this group is already created by the kms-key module.
  # If not yet created, it would fail.
  # We could also create a dependency to that kms-key module, but that may
  # worsen plan/apply time because it would need to run for the kms-key module first.
  group = data.google_cloud_identity_groups.all_groups.groups[
    index(
      data.google_cloud_identity_groups.all_groups.groups[*].group_key[0].id,
      "${each.value.env}-${each.value.service}-techlead@manabie.com",
    )
  ].name

  preferred_member_key {
    id = each.value.manager
  }

  roles {
    name = "MEMBER"
  }
}

# See https://cloud.google.com/kubernetes-engine/docs/how-to/google-groups-rbac#setup-group
resource "google_cloud_identity_group" "gke_security_groups" {
  parent       = "customers/C00ziiz00"
  display_name = "gke-security-groups"
  description  = "Group gke-security-groups@manabie.com managed by Terraform"
  labels       = { "cloudidentity.googleapis.com/groups.discussion_forum" = "" }
  group_key { id = "gke-security-groups@manabie.com" }
}

# The group must have at least 1 owner.
resource "google_cloud_identity_group_membership" "gke_security_owner" {
  group = google_cloud_identity_group.gke_security_groups.id
  preferred_member_key { id = "trieu@manabie.com" }
  roles { name = "MEMBER" }
  roles { name = "OWNER" }
}

# Add all the groups you want to have GKE RBAC enabled here.
resource "google_cloud_identity_group_membership" "gke_security_members" {
  for_each = toset([
    "dev.release@manabie.com",
    "dev.infras@manabie.com",
    "dev.platform@manabie.com",
    "dev.be@manabie.com",
    "devops@manabie.com",
    "tech-func-backend@manabie.com",
    "tech-func-platform@manabie.com",
    "tech-func-automation@manabie.com",
    "tech-squad-platform@manabie.com",
    "tech-squad-automation@manabie.com",
    "tech-squad-release@manabie.com",
    "tech-squad-architecture@manabie.com",
  ])


  group = google_cloud_identity_group.gke_security_groups.id
  preferred_member_key {
    id = each.value
  }
  roles { name = "MEMBER" }
}
