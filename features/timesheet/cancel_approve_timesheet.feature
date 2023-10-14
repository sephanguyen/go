Feature: Cancel Staff Timesheet record

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: Staff cancel approve a valid timesheet
        Given "staff granted role school admin" signin system
        And an existing "APPROVED" timesheet for current staff
        And timesheet has lesson records with "<lesson-statuses>"
        When current staff cancel approve this timesheet
        Then returns "<response-status-code>" status code
        And timesheet status approved changed to submitted "<cancel-approve-status>"

        Examples:
          | signed-in user                    | lesson-statuses               | response-status-code | cancel-approve-status |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully          |

  Scenario Outline: Invalid timesheet status not approved
        Given "<signed-in user>" signin system
        And an existing "<timesheet-status>" timesheet for current staff
        And timesheet has lesson records with "<lesson-statuses>"
        When current staff cancel approve this timesheet
        Then returns "FailedPrecondition" status code
        And timesheet status approved changed to submitted "unsuccessfully"

        Examples:
          | signed-in user                    | lesson-statuses               | timesheet-status |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | DRAFT            |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | CONFIRMED        |
          | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | SUBMITTED        |

  Scenario Outline: Staff cancel approve a timesheet for other staff
    Given an existing "APPROVED" timesheet for other staff "<other-staff-group>"
    When "<signed-in user>" signin system
    And user cancel approve the timesheet for other staff
    Then returns "<resp status-code>" status code
    And timesheet status approved changed to submitted "<cancel-approve-status>"

    Examples:
      | signed-in user                    | resp status-code | cancel-approve-status  | other-staff-group |
      | staff granted role school admin   | OK               | successfully           | staff granted role school admin       |
      | staff granted role teacher        | PermissionDenied | unsuccessfully         | staff granted role school admin       |
      | staff granted role school admin   | OK               | successfully           | staff granted role teacher            |
      | staff granted role teacher        | PermissionDenied | unsuccessfully         | staff granted role teacher            |

  Scenario Outline: Invalid user cancel approve a timesheet for staff
    Given an existing "APPROVED" timesheet for other staff "<other-staff-group>"
    When "<invalid-user>" signin system
    And user cancel approve the timesheet for other staff
    Then returns "<resp status-code>" status code

    Examples:
      | invalid-user    | resp status-code |  other-staff-group                    |
      | unauthenticated | Unauthenticated  |  staff granted role school admin      |
      | parent          | PermissionDenied |  staff granted role teacher           |
