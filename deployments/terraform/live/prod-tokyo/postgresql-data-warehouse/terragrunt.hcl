include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/postgresql"
}

dependency "platforms" {
  config_path = "../platforms"
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/prod-apps.hcl"
}

inputs = {
  pgaudit_enabled = false

  postgresql = {
    project_id = dependency.platforms.outputs.postgresql_project
    instance   = dependency.platforms.outputs.postgresql_instances.data-warehouse
    databases = [
      "kec",
    ]
    users = [
      "prod-kafka-connect@student-coach-e1e95.iam",
      "prod-dwh-kafka-connect@student-coach-e1e95.iam",
    ]
  }

  postgresql_user_permissions = [
    {
      database    = "kec"
      user        = "prod-kafka-connect@student-coach-e1e95.iam"
      schema      = "public"
      object_type = "table"
      privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
    },
    {
      database    = "kec"
      user        = "prod-kafka-connect@student-coach-e1e95.iam"
      schema      = "public"
      object_type = "schema"
      privileges  = ["USAGE"]
    },
    {
      database    = "kec"
      user        = "prod-dwh-kafka-connect@student-coach-e1e95.iam"
      schema      = "public"
      object_type = "table"
      privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
    },
    {
      database    = "kec"
      user        = "prod-dwh-kafka-connect@student-coach-e1e95.iam"
      schema      = "public"
      object_type = "schema"
      privileges  = ["USAGE"]
    },
  ]

  postgresql_statement_timeout = []
  postgresql_bypass_rls_roles  = []
}
