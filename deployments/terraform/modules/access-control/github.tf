locals {
  # array of squad names
  github_squads = distinct(flatten([for member in var.members : [
    for squad in member.squads : squad.name
  ]]))

  # array of function names
  github_functions = distinct(flatten([for member in var.members : [
    for function in member.functions : function.name
  ]]))

  # object keys with all squad / member combinations
  # such as: {  
  #   "squad_platform_ds0nt" = { 
  #     squad = { name = "platform", role = "member" }, 
  #     info = { name = "ds0nt", ... email functions etc } 
  #   },
  #   "squad_platform_bob" = ...
  #   "squad_adobo_joe" = ...
  # }
  github_squads_members = merge(flatten([for member in var.members : {
    for squad in member.squads : "squad_${squad.name}_${member.github.account}" => merge(squad, {
      info = member
    }) if !member.github.disabled
    }
  ])...)

  # same as above but for functions instead of squads
  github_functions_members = merge(flatten([for member in var.members : {
    for function in member.functions : "function_${function.name}_${member.github.account}" => merge(function, {
      info = member
    }) if !member.github.disabled
    }
  ])...)

  # same as github_squads_members but only for members with function "web"
  github_squads_fe_members = {
    for k, squad_member in local.github_squads_members : "${k}_fe" => squad_member
    if contains(squad_member.info.functions[*].name, "web")
  }

  # array of the names of squads with a "web" member
  github_squads_fe = distinct(flatten([
    for k, squad_member in local.github_squads_fe_members :
    [for squad in squad_member.info.squads : squad.name]
  ]))

  # same as github_squads_members but only for members with function "backend"
  github_squads_be_members = {
    for k, squad_member in local.github_squads_members : "${k}_be" => squad_member
    if contains(squad_member.info.functions[*].name, "backend")
  }

  # array of the names of squads with a "backend" member
  github_squads_be = distinct(flatten([
    for k, squad_member in local.github_squads_be_members :
    [for squad in squad_member.info.squads : squad.name]
  ]))

  # same as github_squads_members but only for members with function "mobile"
  github_squads_me_members = {
    for k, squad_member in local.github_squads_members : "${k}_me" => squad_member
    if contains(squad_member.info.functions[*].name, "mobile")
  }

  # array of the names of squads with a "mobile" member
  github_squads_me = distinct(flatten([
    for k, squad_member in local.github_squads_me_members :
    [for squad in squad_member.info.squads : squad.name]
  ]))

  # array of platform members havinng super access level
  github_terraform_approvers = {
    for member in var.members : member.github.account => member
    if 0 < length([
      for function in member.functions : "ok"
      if function.name == "platform" && function.access_level == "super"
    ])
  }
}

resource "github_team" "dev" {
  name    = "dev"
  privacy = "closed"
}

resource "github_team" "functions" {
  for_each = toset(local.github_functions)

  name           = "func-${each.key}"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}

resource "github_team" "squads" {
  for_each = toset(local.github_squads)

  name           = "squad-${each.key}"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}


resource "github_membership" "memberships" {
  for_each = { for member in var.members : member.github.account => member if !member.github.disabled }

  username = each.key
  role     = each.value.github.role
}


resource "github_team_membership" "dev_members" {
  for_each = { for member in var.members : member.github.account => member if !member.github.disabled }

  team_id  = github_team.dev.id
  username = each.value.github.account
  role     = each.value.github.role == "maintainer" ? "maintainer" : "member"
}

resource "github_team_membership" "squads_members" {
  for_each = local.github_squads_members

  team_id  = github_team.squads[each.value.name].id
  username = each.value.info.github.account
  role     = each.value.role == "maintainer" ? "maintainer" : "member"
}

resource "github_team_membership" "functions_members" {
  for_each = local.github_functions_members

  team_id  = github_team.functions[each.value.name].id
  username = each.value.info.github.account
  role     = each.value.role == "member" ? "member" : "maintainer"
}

resource "github_team" "squads_fe" {
  for_each = toset(local.github_squads_fe)

  name           = "squad-${each.key}-fe"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}

resource "github_team_membership" "squads_fe_members" {
  for_each = local.github_squads_fe_members

  team_id  = github_team.squads_fe[each.value.name].id
  username = each.value.info.github.account
  role     = each.value.role == "maintainer" ? "maintainer" : "member"
}

resource "github_team" "terraform_approvers" {
  name           = "terraform-approvers"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}

resource "github_team_membership" "terraform_approvers_membership" {
  for_each = local.github_terraform_approvers

  team_id  = github_team.terraform_approvers.id
  username = each.value.github.account
  role     = each.value.github.account == "nvcnvn" ? "maintainer" : "member"
}

resource "github_team" "squads_be" {
  for_each = toset(local.github_squads_be)

  name           = "squad-${each.key}-be"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}

resource "github_team_membership" "squads_be_members" {
  for_each = local.github_squads_be_members

  team_id  = github_team.squads_be[each.value.name].id
  username = each.value.info.github.account
  role     = each.value.role == "maintainer" ? "maintainer" : "member"
}

resource "github_team" "squads_me" {
  for_each = toset(local.github_squads_me)

  name           = "squad-${each.key}-me"
  privacy        = "closed"
  parent_team_id = github_team.dev.id
}

resource "github_team_membership" "squads_me_members" {
  for_each = local.github_squads_me_members

  team_id  = github_team.squads_me[each.value.name].id
  username = each.value.info.github.account
  role     = each.value.role == "maintainer" ? "maintainer" : "member"
}
