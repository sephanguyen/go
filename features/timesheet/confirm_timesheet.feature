Feature: Confirm Staff Timesheet record

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: Staff confirms a valid timesheet
    Given "<other-staff-group>" signin system
    And staff has an existing "<count-timesheet>" approve timesheet
    And each timesheets has lesson records with "<lesson-statuses>"
    When "staff granted role school admin" staff confirms this timesheet
    Then returns "OK" status code
    And timesheet statuses changed to confirm "successfully"

    Examples:
      | count-timesheet | lesson-statuses               | other-staff-group                    |
      | 1               | CANCELLED-CANCELLED-COMPLETED | staff granted role teacher           |
      | 300             | CANCELLED-COMPLETED-COMPLETED | staff granted role school admin      |

  Scenario Outline: Invalid timesheet status should not be confirmed
    Given "<signed-in user>" signin system
    And an existing "<timesheet-status>" timesheet for current staff
    And timesheet has lesson records with "<lesson-statuses>"
    When current staff confirms this timesheet
    Then returns "Internal" status code
    And timesheet statuses changed to confirm "unsuccessfully"

    Examples:
      | signed-in user                    | lesson-statuses               | timesheet-status |
      | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | DRAFT            |
      | staff granted role school admin   | CANCELLED-COMPLETED-COMPLETED | SUBMITTED        |
      | staff granted role school admin   | CANCELLED-CANCELLED-COMPLETED | CONFIRMED        |

  Scenario Outline: Staff confirm a timesheet for other staff
    Given an existing "APPROVED" timesheet for other staff "<other-staff-group>"
    When "<signed-in user>" signin system
    And user confirms the timesheet for other staff
    Then returns "<resp status-code>" status code
    And timesheet statuses changed to confirm "<confirm-status>"

    Examples:
      | signed-in user                    | resp status-code | confirm-status  | other-staff-group                    |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |

  Scenario Outline: Invalid user confirms a timesheet for staff
    Given an existing "APPROVED" timesheet for other staff "<other-staff-group>"
    When "<invalid-user>" signin system
    And user confirms the timesheet for other staff
    Then returns "<resp status-code>" status code

    Examples:
      | invalid-user    | resp status-code |  other-staff-group                    |
      | unauthenticated | Unauthenticated  |  staff granted role school admin      |
      | parent          | PermissionDenied |  staff granted role teacher           |
