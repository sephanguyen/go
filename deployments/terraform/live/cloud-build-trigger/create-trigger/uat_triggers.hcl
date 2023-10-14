locals {
  ad_hoc_infos = [
    {
      project_id            = "staging-manabie-online"
      org                   = "manabie"
      env                   = "uat"
      service_account_adhoc = "uat-manabie-ad-hoc@staging-manabie-online.iam.gserviceaccount.com"
    },
    {
      project_id            = "staging-manabie-online"
      org                   = "jprep"
      env                   = "uat"
      service_account_adhoc = "uat-jprep-ad-hoc@staging-manabie-online.iam.gserviceaccount.com"
    }
  ]
}