locals {
  ad_hoc_infos = [
    {
      project_id            = "staging-manabie-online"
      org                   = "manabie"
      env                   = "stag"
      service_account_adhoc = "stag-manabie-ad-hoc@staging-manabie-online.iam.gserviceaccount.com"
    },
    {
      project_id            = "staging-manabie-online"
      org                   = "jprep"
      env                   = "stag"
      service_account_adhoc = "stag-jprep-ad-hoc@staging-manabie-online.iam.gserviceaccount.com"
    }
  ]
}