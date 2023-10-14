locals {
  # See custom roles definition in project_roles.hcl.
  pdm = {
    low = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.moderate",
        ]
      }
      uat = {
        can_read_databases = true
        custom_roles = [
          "customroles.pdm.low",
        ]
      }
    }
    moderate = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.moderate",
        ]
      }
      prod = {
        custom_roles = [
          "customroles.pdm.low",
        ]
      }
    }
    high = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.high",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.high",
        ]
      }
      prod = {
        can_read_databases = true
        custom_roles = [
          "customroles.pdm.moderate",
        ]
      }
    }
    super = {
      stag = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.super",
        ]
      }
      uat = {
        can_read_databases  = true
        can_write_databases = true
        custom_roles = [
          "customroles.pdm.super",
        ]
      }
      prod = {
        can_read_databases  = true
        can_write_databases = false
        custom_roles = [
          "customroles.pdm.super",
        ]
      }
    }
  }
}
