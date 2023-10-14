terraform {
  required_version = "~> 1.4.0"
}

locals {
  member_function_level = flatten([
    for function, levels in var.member_by_access_level : [
      for level, members in levels : [
        for member in members : {
          member   = member
          function = function
          level    = level
        }
      ]
    ]
  ])

  member_function_level_role = flatten([
    for m in local.member_function_level : [
      for role in try(var.role_by_access_level[m.function][m.level][var.env].custom_roles, []) : {
        member         = m.member
        function       = m.function
        level          = m.level
        custom_role_id = role
      }
    ]
  ])

  techlead_member_role = flatten([
    for m in var.techleads : [
      for role in try(var.techlead_roles[var.env].custom_roles, []) : {
        member         = m
        custom_role_id = role
      }
    ]
  ])
}

resource "google_project_iam_member" "member_roles" {
  for_each = {
    for m in local.member_function_level_role :
    "${m.member}.${m.custom_role_id}" => m
  }

  project = var.project_id
  role    = "projects/${var.project_id}/roles/${each.value.custom_role_id}"
  member  = "user:${each.value.member}"
}


resource "google_project_iam_member" "techlead_roles" {
  for_each = {
    for m in local.techlead_member_role :
    "${m.member}.${m.custom_role_id}" => m
  }

  project = var.project_id
  role    = "projects/${var.project_id}/roles/${each.value.custom_role_id}"
  member  = "user:${each.value.member}"
}

output "member_function_level_role" {
  value = local.member_function_level_role
}
