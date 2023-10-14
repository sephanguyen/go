locals {
  # See custom roles definition in project_roles.hcl.
  platform = {
    low = {
      stag = {
        can_read_databases = true
        custom_roles = [
          "customroles.platform.moderate",
        ]
      }
      uat = {
        custom_roles = [
          "customroles.platform.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.high",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.platform.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.platform.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.platform.high",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.platform.super",
        ]
      }
    }
  }
}
