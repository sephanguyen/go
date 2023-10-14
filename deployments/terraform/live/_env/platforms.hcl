locals {
  database_flags = [
    {
      name  = "log_min_duration_statement"
      value = "250"
    },
    {
      name  = "log_checkpoints"
      value = "on"
    },
    {
      name  = "log_connections"
      value = "on"
    },
    {
      name  = "log_disconnections"
      value = "on"
    },
    {
      name  = "autovacuum"
      value = "on"
    },
    {
      name  = "cloudsql.iam_authentication"
      value = "on"
    },
    {
      name  = "cloudsql.logical_decoding"
      value = "on"
    }
  ]

  insights_config = {
    query_string_length     = 1024
    record_application_tags = false
    record_client_address   = false
  }
}
