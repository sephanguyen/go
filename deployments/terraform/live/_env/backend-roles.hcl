locals {
  # See custom roles definition in project_roles.hcl.
  backend = {
    low = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.moderate",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.backend.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.backend.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.backend.moderate",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.backend.super",
        ]
      }
    }
  }
}
