include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../../modules/change-db-type"
}

inputs = {
  function_name = "change-db-type"
  entry_point   = "ProcessPubSub"
  // Slack channel: #stag-uat-monitoring
  slack_webhook = "https://hooks.slack.com/services/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"
  topic_name    = "change-db-type-topic"
  scheduler_jobs = [
    {
      name : "upgrade-db-type-Monday",
      description : "6 CPUs for DB on Monday (from 6:30AM +7)",
      schedule : "30 6 * * 1",
      data : "{\"instance\": \"manabie-common-88e1ee71\",\"project\": \"staging-manabie-online\",\"instance_type\": \"db-custom-6-23040\"}",
    },
    {
      name : "upgrade-db-type-weekdays",
      description : "4 CPUs for DB on weekdays (from Monday 7:30PM +7)",
      schedule : "30 19 * * 1",
      data : "{\"instance\": \"manabie-common-88e1ee71\",\"project\": \"staging-manabie-online\",\"instance_type\": \"db-custom-4-15360\"}",
    },
    {
      name : "downgrade-db-type-weekend",
      description : "Downsize DB -> 2 CPUs on Saturday 0AM",
      schedule : "0 0 * * 6",
      data : "{\"instance\": \"manabie-common-88e1ee71\",\"project\": \"staging-manabie-online\",\"instance_type\": \"db-custom-2-7680\"}",
    },
  ]
}
