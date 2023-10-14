Feature: Send lesson lock event

  Scenario Outline: Send lesson lock event publish successfully
      Given "<other-staff-group>" signin system
      And staff has an existing "<count-timesheet>" submitted timesheet
      And each timesheets has lesson records with "<lesson-statuses>"
      When timesheet send event lock lesson
      Then timesheet event lock lesson published successfully

      Examples:
        | count-timesheet | lesson-statuses               | response-status-code | approve-status | other-staff-group           |
        | 100             | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully   | staff granted role teacher  |