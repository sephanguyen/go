locals {
  access_levels = distinct(compact(flatten(var.members[*].functions[*].access_level)))
  squad_roles   = distinct(compact(flatten(var.members[*].squads[*].role)))
}

output "members_by_access_level" {
  value = {
    for function in local.github_functions :
    function => {
      for level in local.access_levels :
      level => flatten([
        for member in var.members : [
          for f in member.functions : member.email
          if f.name == function && f.access_level == level
        ]
      ])
    }
  }
}

output "techleads" {
  value = distinct(flatten([
    for member in var.members : [
      for member_function in member.functions :
      member.email if member_function.name == "techlead"
    ]
  ]))
}
