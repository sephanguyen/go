locals {
  # See custom roles definition in project_roles.hcl.
  data = {
    low = {
      stag = {
        can_read_databases = true
        custom_roles = [
          "customroles.data.moderate",
        ]
      }
      uat = {
        custom_roles = [
          "customroles.data.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.high",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.data.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.data.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.data.moderate",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.data.super",
        ]
      }
    }
  }
}
