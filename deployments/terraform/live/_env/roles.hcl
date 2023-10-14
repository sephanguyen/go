locals {
  platform_roles = read_terragrunt_config("./platform-roles.hcl").locals
  data_roles     = read_terragrunt_config("./data-roles.hcl").locals
  backend_roles  = read_terragrunt_config("./backend-roles.hcl").locals
  web_roles      = read_terragrunt_config("./web-roles.hcl").locals
  mobile_roles   = read_terragrunt_config("./mobile-roles.hcl").locals
  # pdm_roles      = read_terragrunt_config("./pdm-roles.hcl").locals

  role_by_access_level = {
    platform = local.platform_roles.platform
    data     = local.data_roles.data
    backend  = local.backend_roles.backend
    web      = local.web_roles.web
    mobile   = local.mobile_roles.mobile
    # pdm      = local.pdm_roles.pdm
  }

  # Roles to be granted to "techlead" function
  techlead_roles = {
    stag = {
      custom_roles = [
        "customroles.cloudbuild.approval",
      ]
    }
    uat = {
      custom_roles = [
        "customroles.cloudbuild.approval",
      ]
    }
    prod = {
      custom_roles = [
        "customroles.cloudbuild.approval",
      ]
    }
  }
}
