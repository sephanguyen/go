include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/access-control"
}

locals {
  token = yamldecode(sops_decrypt_file(find_in_parent_folders("token.enc.yaml")))

  bots = [
    {
      name  = "botmanabie"
      email = "botmanabie@manabie.com"
      google_group = {
        disabled = true
      }
      jira = {
        disabled = true
      }
      slack = {
        disabled = true
      }
      github = {
        account = "botmanabie"
        role    = "member"
      }
      squads = [
        {
          name = "platform"
          role = "member"
        },
      ]
      functions = [
        {
          name = "web"
          role = "member"
        },
      ]
    },
    {
      name  = "manaops"
      email = "devops@manabie.com"
      google_group = {
        disabled = true
      }
      slack = {
        disabled = true
      }
      github = {
        account = "manaops"
        role    = "admin"
      }
      squads = [
        {
          name = "admin"
          role = "maintainer"
        },
        {
          name = "platform"
          role = "member"
        },
      ]
      functions = [
        {
          name = "platform"
          role = "member"
        },
      ]
    },
    {
      name  = "lhtrieu87"
      email = "trieu@manabie.com"
      slack = {
        disabled       = false
        email_override = "lhtrieu87@gmail.com"
      }

      github = {
        account = "lhtrieu87"
        role    = "admin"
      }
      squads = [
        {
          name = "admin"
          role = "manager"
        },
        {
          name = "release"
          role = "manager"
        },
        {
          name = "ddd"
          role = "manager"
        },
      ]
      functions = []
    },
  ]
  communication             = read_terragrunt_config("communication_members.hcl").locals
  adobo                     = read_terragrunt_config("adobo_members.hcl").locals
  lesson                    = read_terragrunt_config("lesson_members.hcl").locals
  platform                  = read_terragrunt_config("platform_members.hcl").locals
  syllabus                  = read_terragrunt_config("syllabus_members.hcl").locals
  user_management           = read_terragrunt_config("user_management_members.hcl").locals
  payment                   = read_terragrunt_config("payment_members.hcl").locals
  cse                       = read_terragrunt_config("cse_members.hcl").locals
  data                      = read_terragrunt_config("data_members.hcl").locals
  architecture              = read_terragrunt_config("architecture_members.hcl").locals
  calendar                  = read_terragrunt_config("calendar_members.hcl").locals
  timesheet                 = read_terragrunt_config("timesheet_members.hcl").locals
  internship                = read_terragrunt_config("internship_members.hcl").locals
  virtual_classroom         = read_terragrunt_config("virtual_classroom_members.hcl").locals
  auth                      = read_terragrunt_config("auth_members.hcl").locals
  hermes                    = read_terragrunt_config("hermes_members.hcl").locals
  order_management          = read_terragrunt_config("order_management_members.hcl").locals
  student_course_management = read_terragrunt_config("student_course_management_members.hcl").locals

  # Import the global service defintions
  owners                   = yamldecode(file("${get_repo_root()}/deployments/decl/owners.yaml")).owners
  stag_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/stag-defs.yaml"))
  uat_service_definitions  = yamldecode(file("${get_repo_root()}/deployments/decl/uat-defs.yaml"))
  prod_service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/prod-defs.yaml"))

  service_ownerships = concat(
    [
      for service in local.stag_service_definitions : {
        service = service.name, env = "stag", squad = try(local.owners[service.name], "platform")
      } if !try(service.disable_iam, false)
    ],
    [
      for service in local.uat_service_definitions : {
        service = service.name, env = "uat", squad = try(local.owners[service.name], "platform")
      } if !try(service.disable_iam, false)
    ],
    [
      for service in local.prod_service_definitions : {
        service = service.name, env = "prod", squad = try(local.owners[service.name], "platform")
      } if !try(service.disable_iam, false)
    ],
  )
}

inputs = {
  project_id   = "student-coach-e1e95"
  github_token = local.token.github
  slack_token  = local.token.slack
  jira_token   = local.token.jira
  members = concat(
    local.bots,
    local.communication.members,
    local.adobo.members,
    local.lesson.members,
    local.platform.members,
    local.syllabus.members,
    local.user_management.members,
    local.payment.members,
    local.cse.members,
    local.data.members,
    local.architecture.members,
    local.calendar.members,
    local.timesheet.members,
    local.internship.members,
    local.virtual_classroom.members,
    local.auth.members,
    local.hermes.members,
    local.order_management.members,
    local.student_course_management.members,
  )
  service_ownerships = local.service_ownerships
}
