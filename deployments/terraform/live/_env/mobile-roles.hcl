locals {
  # See custom roles definition in project_roles.hcl.
  mobile = {
    low = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.moderate",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.mobile.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.mobile.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.mobile.moderate",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.mobile.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = false
        custom_roles = [
          "customroles.mobile.super",
        ]
      }
    }
  }
}
