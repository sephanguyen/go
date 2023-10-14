# In slack.tf we create slack handles.
# for example:
#   @tech-squad-platform
#   @tech-func-web
#   @tech-squad-adobo-web

data "slack_user" "users" {
  for_each = { for member in var.members : member.email => member if !member.slack.disabled }

  email = coalesce(each.value.slack.email_override, each.key)
}

locals {
  # used for handles like @tech-func-web
  slack_function_usergroups = { for function in local.github_functions :
    function => flatten([
      for member in var.members : [
        for func in member.functions :
        data.slack_user.users[member.email].id if func.name == function
      ]
      if !member.slack.disabled
    ])
  }

  # used for handles like @tech-squad-adobo
  slack_squad_usergroups = { for squad in local.github_squads :
    squad => flatten([
      for member in var.members : [
        for func in member.squads :
        data.slack_user.users[member.email].id if func.name == squad
      ]
      if !member.slack.disabled
    ])
  }

  # used for handles like @tech-squad-adobo-web
  slack_squad_func_usergroups = merge(flatten([
    for squad in local.github_squads : {
      for function in local.github_functions : "${squad}-${function}" => distinct(flatten([
        for member in var.members : [
          for mem_squad in member.squads : [
            for mem_func in member.functions :
            data.slack_user.users[member.email].id if mem_squad.name == squad && mem_func.name == function
          ]
        ]
        if !member.slack.disabled
      ]))
    }
  ])...)
}

resource "slack_usergroup" "devs" {
  name        = "tech-devs"
  handle      = "tech-devs"
  description = "tech-devs managed by terraform script"
  users       = values(data.slack_user.users)[*].id

  lifecycle {
    ignore_changes = [
      channels,
    ]
  }
}

resource "slack_usergroup" "function_usergroups" {
  for_each = local.slack_function_usergroups

  name        = "tech-func-${each.key}"
  handle      = "tech-func-${each.key}"
  description = "tech-func-${each.key} managed by terraform script"
  users       = each.value

  lifecycle {
    ignore_changes = [
      channels,
    ]
  }
}

resource "slack_usergroup" "squad_usergroups" {
  for_each = local.slack_squad_usergroups

  name        = "tech-squad-${each.key}"
  handle      = "tech-squad-${each.key}"
  description = "tech-squad-${each.key} managed by terraform script"
  users       = each.value

  lifecycle {
    ignore_changes = [
      channels,
    ]
  }
}

resource "slack_usergroup" "squad_func_usergroups" {
  for_each = {
    for key, val in local.slack_squad_func_usergroups :
    key => val if length(val) > 0
  }

  name        = "tech-squad-${each.key}"
  handle      = "tech-squad-${each.key}"
  description = "tech-squad-${each.key} managed by terraform script"
  users       = each.value

  lifecycle {
    ignore_changes = [
      channels,
    ]
  }
}
