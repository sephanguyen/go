locals {
  runs_on_project_id = "student-coach-e1e95"
  hasura_host        = "admin.prod.ga.manabie.io"
  hasura_paths = [
    "/healthz",
    "/calendar/healthz",
    "/entryexitmgmt/healthz",
    "/eureka/healthz",
    "/fatima/healthz",
    "/invoicemgmt/healthz",
    "/lessonmgmt/healthz",
    "/mastermgmt/healthz",
    "/timesheet/healthz",
  ]
  
  https_check = {
    cloudconvert-job-event = {
      name = "CloudConvert job event"
      host = "web-api.prod.ga.manabie.io"
      port = 443
      request_method = "POST"
      content_type = "URL_ENCODED"
      content_matchers = {
        content = "{\"error\":\"signature is not match\"}"
        matcher = "CONTAINS_STRING"
      }
      accepted_response_status_codes = {
          status_value = 400 #Bad request
      }
      paths = [
        "/cloud-convert/job-events"
      ]
    }
  }
}
