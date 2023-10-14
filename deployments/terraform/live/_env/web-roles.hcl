locals {
  # See custom roles definition in project_roles.hcl.
  web = {
    low = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.moderate",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.web.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.web.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.web.moderate",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.web.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = false
        custom_roles = [
          "customroles.web.super",
        ]
      }
    }
  }
}
