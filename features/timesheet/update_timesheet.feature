Feature: Update timesheet
  Background:
    Given have timesheet configuration is on

  Scenario Outline: Staff update a timesheet for ourselves
    Given "<signed-in user>" signin system
    And new updated timesheet data with "<timesheet-status>" status for current staff
    When user update a timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code | timesheet-status |
      | staff granted role school admin   | OK               | DRAFT            |
      | staff granted role teacher        | OK               | DRAFT            |
      | staff granted role school admin   | OK               | SUBMITTED        |
      | staff granted role teacher        | PermissionDenied | SUBMITTED        |
      | staff granted role school admin   | PermissionDenied | APPROVED         |
      | staff granted role teacher        | PermissionDenied | APPROVED         |
      | staff granted role school admin   | PermissionDenied | CONFIRMED        |
      | staff granted role teacher        | PermissionDenied | CONFIRMED        |

  Scenario Outline: Staff update a timesheet for other staff
    Given new updated timesheet data with "<timesheet-status>" status for other staff "<other-staff-group>"
    When "<signed-in user>" signin system
    And user update a timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code | timesheet-status | other-staff-group                    |
      | staff granted role school admin   | OK               | DRAFT            | staff granted role school admin      |
      | staff granted role school admin   | OK               | DRAFT            | staff granted role teacher           |
      | staff granted role school admin   | OK               | SUBMITTED        | staff granted role school admin      |
      | staff granted role school admin   | OK               | SUBMITTED        | staff granted role teacher           |
      | staff granted role school admin   | PermissionDenied | APPROVED         | staff granted role school admin      |
      | staff granted role school admin   | PermissionDenied | APPROVED         | staff granted role teacher           |
      | staff granted role school admin   | PermissionDenied | CONFIRMED        | staff granted role school admin      |
      | staff granted role school admin   | PermissionDenied | CONFIRMED        | staff granted role teacher           |
      | staff granted role teacher        | PermissionDenied | DRAFT            | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | DRAFT            | staff granted role teacher           |
      | staff granted role teacher        | PermissionDenied | SUBMITTED        | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | SUBMITTED        | staff granted role teacher           |
      | staff granted role teacher        | PermissionDenied | APPROVED         | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | APPROVED         | staff granted role teacher           |
      | staff granted role teacher        | PermissionDenied | CONFIRMED        | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | CONFIRMED        | staff granted role teacher           |

  Scenario Outline: Invalid user update a timesheet for ourselves
    Given "<invalid user>" signin system
    When user update a timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | invalid user    | resp status-code |
      | unauthenticated | Unauthenticated  |
      | parent          | PermissionDenied |

  Scenario Outline: Invalid user update a timesheet for other staff
    Given "<invalid user>" signin system
    When user update a timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | invalid user    | resp status-code |
      | unauthenticated | Unauthenticated  |
      | parent          | PermissionDenied |
