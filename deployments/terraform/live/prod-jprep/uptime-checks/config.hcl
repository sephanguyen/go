locals {
  runs_on_project_id = "student-coach-e1e95"
  hasura_host        = "admin.prod.jprep.manabie.io"
  hasura_paths = [
    "/healthz",
    "/eureka/healthz",
    "/fatima/healthz",
    "/invoicemgmt/healthz",
    "/lessonmgmt/healthz",
    "/mastermgmt/healthz",
    "/timesheet/healthz",
  ]
  https_check = {
    JPREP-healthz-integrate-apis = {
      name = "JPREP healthz integrate APIs"
      host = "web-api.prod.jprep.manabie.io"
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
      body = "ewogICAgInVybCI6ICJodHRwczovL3dlYi1hcGkuc3RhZ2luZy5qcHJlcC5tYW5hYmllLmlvLyIsCiAgICAiZXJyb3JfY29kZSI6IDQwMCwKICAgICJjb250ZW50X21hdGNoIjogInNpZ25hdHVyZSBpcyBub3QgbWF0Y2giLAogICAgInBhdGhzIjogWwogICAgICAgICJqcHJlcC91c2VyLXJlZ2lzdHJhdGlvbiIsCiAgICAgICAgImpwcmVwL3VzZXItY291cnNlIiwKICAgICAgICAianByZXAvbWFzdGVyLXJlZ2lzdHJhdGlvbiIKICAgIF0KfQ=="
    }
    JPREP-healthz-integrate-apis-old-ports = {
      name = "JPREP healthz integrate APIs orl port"
      host = "web-api.prod.jprep.manabie.io"
      port = 31400
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
      body = "ewogICAgInVybCI6ICJodHRwczovL3dlYi1hcGkuc3RhZ2luZy5qcHJlcC5tYW5hYmllLmlvLyIsCiAgICAiZXJyb3JfY29kZSI6IDQwMCwKICAgICJjb250ZW50X21hdGNoIjogInNpZ25hdHVyZSBpcyBub3QgbWF0Y2giLAogICAgInBhdGhzIjogWwogICAgICAgICJqcHJlcC91c2VyLXJlZ2lzdHJhdGlvbiIsCiAgICAgICAgImpwcmVwL3VzZXItY291cnNlIiwKICAgICAgICAianByZXAvbWFzdGVyLXJlZ2lzdHJhdGlvbiIKICAgIF0KfQ=="
    }
    JPREP-slackbot-orl-port = {
      name = "PROD JPREP slackbot orl port"
      host = "web-api.prod.jprep.manabie.io"
      port = 31400
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
    JPREP-slackbot = {
      name = "PROD JPREP slackbot"
      host = "web-api.prod.jprep.manabie.io"
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
      host = "web-api.prod.jprep.manabie.io"
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
    cloudconvert-job-event-orl-port = {
      name = "CloudConvert job event orl port"
      host = "web-api.prod.jprep.manabie.io"
      port = 31400
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
