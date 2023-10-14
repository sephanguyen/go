variable "ad_hoc_infos" {
  type = list(object({
    project_id            = string
    org                   = string
    env                   = string
    service_account_adhoc = string
  }))
}
