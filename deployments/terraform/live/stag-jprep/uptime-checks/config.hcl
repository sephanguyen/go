locals {
  hasura_host = "admin.staging.jprep.manabie.io"
  hasura_paths = [
    "/healthz",
    "/auth/healthz",
    "/eureka/healthz",
    "/fatima/healthz",
    "/lessonmgmt/healthz",
    "/invoicemgmt/healthz",
    "/timesheet/healthz",
    "/mastermgmt/healthz",
  ]

  https_check = {
    JPREP-healthz-integrate-apis = {
      name = "JPREP healthz integrate APIs"
      host = "web-api.staging.jprep.manabie.io"
      port = 443
      request_method = "POST"
      content_type = "URL_ENCODED"
      content_matchers = {
        content = "{\"message\":\"OK\"}"
        matcher = "CONTAINS_STRING"
      }
      accepted_response_status_codes = {
          status_value = 200 #OK
      }
      paths = [
        "/healthcheck/jprep", 
      ]
      # {
      #   "url": "https://web-api.staging.jprep.manabie.io/",
      #   "error_code": 400,
      #   "content_match": "signature is not match",
      #   "paths": [
      #       "jprep/user-registration",
      #       "jprep/user-course",
      #       "jprep/master-registration"
      #   ]
      # }
      body = "ewogICAgInVybCI6ICJodHRwczovL3dlYi1hcGkuc3RhZ2luZy5qcHJlcC5tYW5hYmllLmlvLyIsCiAgICAiZXJyb3JfY29kZSI6IDQwMCwKICAgICJjb250ZW50X21hdGNoIjogInNpZ25hdHVyZSBpcyBub3QgbWF0Y2giLAogICAgInBhdGhzIjogWwogICAgICAgICJqcHJlcC91c2VyLXJlZ2lzdHJhdGlvbiIsCiAgICAgICAgImpwcmVwL3VzZXItY291cnNlIiwKICAgICAgICAianByZXAvbWFzdGVyLXJlZ2lzdHJhdGlvbiIKICAgIF0KfQ=="
    }
    JPREP-slackbot = {
      name = "JPREP slackbot APIs"
      host = "web-api.staging.jprep.manabie.io"
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
        "/jprep/partner-log", 
      ]
    }
    cloudconvert-job-event = {
      name = "CloudConvert job event"
      host = "web-api.staging.jprep.manabie.io"
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
