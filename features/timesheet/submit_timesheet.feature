Feature: Submit Staff Timesheet record
  
  Background:
    Given have timesheet configuration is on

  Scenario Outline: Staff submits a valid timesheet with no lesson record successfully
      Given "<signed-in user>" signin system
      And an existing "DRAFT" timesheet with date "<timesheet-date>" for current staff
      When current staff submits this timesheet
      Then returns "OK" status code
      And timesheet status changed to submitted "successfully"

      Examples:
        | signed-in user                    | timesheet-date |
        | staff granted role school admin   | TODAY          |
        | staff granted role teacher        | TODAY          |
        | staff granted role school admin   | YESTERDAY      |
        | staff granted role teacher        | YESTERDAY      |


  Scenario Outline: Staff submits an invalid future timesheet date
      Given "<signed-in user>" signin system
      And an existing "DRAFT" timesheet with date "<timesheet-date>" for current staff
      When current staff submits this timesheet
      Then returns "FailedPrecondition" status code
      And timesheet status changed to submitted "unsuccessfully"

      Examples:
        | signed-in user                    | timesheet-date     |
        | staff granted role school admin   | TOMORROW           |
        | staff granted role teacher        | 5DAYS FROM TODAY   |
        | staff granted role teacher        | 2MONTHS FROM TODAY |

  Scenario Outline: Staff submits a timesheet with lesson status
      Given "<signed-in user>" signin system
      And an existing "DRAFT" timesheet with date "<timesheet-date>" for current staff
      And timesheet has lesson records with "<lesson-statuses>"
      When current staff submits this timesheet
      Then returns "<response-status-code>" status code
      And timesheet status changed to submitted "<submit-status>"

      Examples:
        | signed-in user                    | timesheet-date | lesson-statuses               | response-status-code | submit-status  |
        | staff granted role school admin   | TODAY          | CANCELLED-PUBLISHED-COMPLETED | FailedPrecondition   | unsuccessfully |
        | staff granted role teacher        | TODAY          | CANCELLED-PUBLISHED-PUBLISHED | FailedPrecondition   | unsuccessfully |
        | staff granted role school admin   | YESTERDAY      | CANCELLED-PUBLISHED-CANCELLED | FailedPrecondition   | unsuccessfully |
        | staff granted role teacher        | YESTERDAY      | COMPLETED-COMPLETED-PUBLISHED | FailedPrecondition   | unsuccessfully |
        | staff granted role school admin   | TODAY          | CANCELLED-CANCELLED-COMPLETED | OK                   | successfully   |
        | staff granted role teacher        | TODAY          | CANCELLED-COMPLETED-COMPLETED | OK                   | successfully   |


  Scenario Outline: Staff submits a timesheet with invalid timesheet status
    Given "<signed-in user>" signin system
    And an existing "<timesheet-status>" timesheet for current staff
    When current staff submits this timesheet
    Then returns "FailedPrecondition" status code

    Examples:
      | signed-in user                    | timesheet-status |
      | staff granted role school admin   | APPROVED         |
      | staff granted role teacher        | APPROVED         |
      | staff granted role school admin   | SUBMITTED        |
      | staff granted role teacher        | SUBMITTED        |
      | staff granted role school admin   | CONFIRMED        |
      | staff granted role teacher        | CONFIRMED        |

  Scenario Outline: Staff submits a timesheet for other staff
    Given an existing "DRAFT" timesheet for other staff "<other-staff-group>"
    When "<signed-in user>" signin system
    And user submits the timesheet for other staff
    Then returns "<resp status-code>" status code
    And timesheet status changed to submitted "<submit-status>"

    Examples:
      | signed-in user                    | resp status-code | submit-status   | other-staff-group                    |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |

  Scenario Outline: Invalid user submits a timesheet for staff
    Given an existing "DRAFT" timesheet for other staff "<other-staff-group>"
    When "<invalid-user>" signin system
    And user submits the timesheet for other staff
    Then returns "<resp status-code>" status code

    Examples:
      | invalid-user    | resp status-code |  other-staff-group                    |
      | unauthenticated | Unauthenticated  |  staff granted role school admin      |
      | parent          | PermissionDenied |  staff granted role teacher           |
