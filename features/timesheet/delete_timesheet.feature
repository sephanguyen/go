Feature: Delete Staff Timesheet record

  Background:
    Given have timesheet configuration is on

  Scenario Outline: Staff deletes a draft timesheet successfully
    Given "<signed-in user>" signin system
    And an existing "DRAFT" timesheet for current staff
    And timesheet has "<other-working-hours-count>" other working hours records
    And timesheet has "<transport-expense-count>" transport expenses records
    When current staff deletes this timesheet
    Then returns "OK" status code
    And timesheet is deleted "successfully"
    And timesheet other working hours records is deleted "<other-working-hours-deleted>"
    And timesheet transport expenses records is deleted "<transport-expenses-deleted>"

    Examples:
      | signed-in user                    | other-working-hours-count  | other-working-hours-deleted | transport-expense-count  | transport-expenses-deleted |
      | staff granted role school admin   | 0                          | unsuccessfully              | 0                        | unsuccessfully             |   
      | staff granted role teacher        | 0                          | unsuccessfully              | 0                        | unsuccessfully             |
      | staff granted role school admin   | 3                          | successfully                | 3                        | successfully               |
      | staff granted role teacher        | 5                          | successfully                | 5                        | successfully               |
  
  Scenario Outline: Approver/Confirmer deletes a submitted timesheet successfully
    Given "<signed-in user>" signin system
    And an existing "SUBMITTED" timesheet for current staff
    When current staff deletes this timesheet
    Then returns "OK" status code
    And timesheet is deleted "successfully"

    Examples:
      | signed-in user                    |
      | staff granted role school admin   |
      | staff granted role school admin   |

  Scenario Outline: Staff deletes a timesheet with invalid timesheet status
    Given "<signed-in user>" signin system
    And an existing "<timesheet-status>" timesheet for current staff
    When current staff deletes this timesheet
    Then returns "FailedPrecondition" status code

    Examples:
      | signed-in user                    | timesheet-status |
      | staff granted role school admin   | APPROVED         |
      | staff granted role teacher        | APPROVED         |
      | staff granted role school admin   | CONFIRMED        |
      | staff granted role teacher        | CONFIRMED        |

  Scenario Outline: Requester deletes a submitted timesheet
    Given "<signed-in user>" signin system
    And an existing "SUBMITTED" timesheet for current staff
    When current staff deletes this timesheet
    Then returns "PermissionDenied" status code

    Examples:
      | signed-in user                    |
      | staff granted role teacher        |

   Scenario Outline: Staff deletes a timesheet for other staff
    Given an existing "DRAFT" timesheet for other staff "<other-staff-group>"
    And timesheet has "<other-working-hours-count>" other working hours records
    And timesheet has "<transport-expense-count>" transport expenses records
    When "<signed-in user>" signin system
    And user deletes the timesheet for other staff
    Then returns "<resp status-code>" status code
    And timesheet is deleted "<timesheet-deleted>"
    And timesheet other working hours records is deleted "<other-working-hours-deleted>"
    And timesheet transport expenses records is deleted "<transport-expenses-deleted>"

    Examples:
      | signed-in user                    | resp status-code | timesheet-deleted | other-staff-group                    | other-working-hours-count  | other-working-hours-deleted | transport-expense-count  | transport-expenses-deleted |
      | staff granted role school admin   | OK               | successfully      | staff granted role school admin      | 0                          | unsuccessfully              | 0                        | unsuccessfully             |
      | staff granted role teacher        | PermissionDenied | unsuccessfully    | staff granted role teacher           | 0                          | unsuccessfully              | 0                        | unsuccessfully             |
      | staff granted role school admin   | OK               | successfully      | staff granted role school admin      | 3                          | successfully                | 3                        | successfully               |
      | staff granted role teacher        | PermissionDenied | unsuccessfully    | staff granted role teacher           | 5                          | unsuccessfully              | 5                        | unsuccessfully             |

  Scenario Outline: Invalid user deletes a timesheet for staff
    Given an existing "DRAFT" timesheet for other staff "<other-staff-group>"
    When "<invalid-user>" signin system
    And user deletes the timesheet for other staff
    Then returns "<resp status-code>" status code

    Examples:
      | invalid-user    | resp status-code |  other-staff-group                   |
      | unauthenticated | Unauthenticated  |  staff granted role school admin     |
      | parent          | PermissionDenied |  staff granted role teacher          |


  Scenario Outline: Staff deletes a valid timesheet with lesson record
    Given "<signed-in user>" signin system
    And an existing "DRAFT" timesheet for current staff
    And timesheet has "<other-working-hours-count>" other working hours records
    And timesheet has lesson records
    When current staff deletes this timesheet
    Then returns "FailedPrecondition" status code
    And timesheet is deleted "unsuccessfully"
    And timesheet other working hours records is deleted "unsuccessfully"

    Examples:
      | signed-in user                    | other-working-hours-count |
      | staff granted role school admin   | 0                         |
      | staff granted role teacher        | 0                         |
      | staff granted role school admin   | 3                         |
      | staff granted role teacher        | 5                         |
