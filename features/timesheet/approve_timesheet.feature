Feature: Approve Staff Timesheet record

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: Staff approves a valid timesheet
        Given "staff granted role school admin" signin system
        And an existing "SUBMITTED" timesheet for current staff
        And timesheet has lesson records with "<lesson-statuses>"
        When current staff approves this timesheet
        Then returns "<response-status-code>" status code
        And timesheet status changed to approve "<approve-status>"

        Examples:
          | lesson-statuses               | response-status-code | approve-status |
          | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully   |

  Scenario Outline: Staff approves a multiple valid timesheet
        Given "<other-staff-group>" signin system
        And staff has an existing "<count-timesheet>" submitted timesheet
        And each timesheets has lesson records with "<lesson-statuses>"
        When "staff granted role school admin" staff approves this timesheet
        Then returns "<response-status-code>" status code
        And timesheet status changed to approve "<approve-status>"

        Examples:
          | count-timesheet | lesson-statuses               | response-status-code | approve-status | other-staff-group                    |
          | 50              | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully   | staff granted role teacher           |
          | 100             | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully   | staff granted role school admin      |

  Scenario Outline: Staff approves a valid timesheet with invalid lession status
        Given "staff granted role school admin" signin system
        And an existing "SUBMITTED" timesheet for current staff
        And timesheet has lesson records with "<lesson-statuses>"
        When current staff approves this timesheet
        Then returns "<response-status-code>" status code
        And timesheet status changed to approve "<approve-status>"

        Examples:
          | response-status-code| lesson-statuses     | approve-status |
          | FailedPrecondition  | PUBLISHED           | unsuccessfully |

  Scenario Outline: Invalid timesheet status should not be approved
        Given "<signed-in user>" signin system
        And an existing "<timesheet-status>" timesheet for current staff
        And timesheet has lesson records with "<lesson-statuses>"
        When current staff approves this timesheet
        Then returns "Internal" status code
        And timesheet status changed to approve "unsuccessfully"

        Examples:
          | signed-in user                    | lesson-statuses               | timesheet-status |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | DRAFT            |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | CONFIRMED        |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | APPROVED         |

  Scenario Outline: Staff approves a timesheet for other staff
      Given an existing "SUBMITTED" timesheet for other staff "<other-staff-group>"
      When "<signed-in user>" signin system
      And user approves the timesheet for other staff
      Then returns "<resp status-code>" status code
      And timesheet status changed to approve "<approve-status>"

      Examples:
        | signed-in user                    | resp status-code | approve-status  | other-staff-group                    |
        | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
        | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |
        | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
        | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |

  Scenario Outline: Invalid user approves a timesheet for staff
    Given an existing "SUBMITTED" timesheet for other staff "<other-staff-group>"
    When "<invalid-user>" signin system
    And user approves the timesheet for other staff
    Then returns "<resp status-code>" status code

    Examples:
      | invalid-user    | resp status-code |  other-staff-group                    |
      | unauthenticated | Unauthenticated  |  staff granted role school admin      |
      | parent          | PermissionDenied |  staff granted role teacher           |
