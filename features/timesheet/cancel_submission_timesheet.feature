Feature: Cancel Submit Staff Timesheet record

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: Staff cancel submit a valid timesheet
      Given "<signed-in user>" signin system
      And an existing "SUBMITTED" timesheet for current staff
      When current staff cancel submits this timesheet
      Then returns "OK" status code
      And timesheet status changed to draft "successfully"

      Examples:
        | signed-in user                    |
        | staff granted role school admin   |
        | staff granted role teacher        |
        | staff granted role school admin   |
        | staff granted role teacher        |
        
  Scenario Outline: Staff cancel submit a timesheet with invalid timesheet status
    Given "<signed-in user>" signin system
    And an existing "<timesheet-status>" timesheet for current staff
    When current staff cancel submits this timesheet
    Then returns "FailedPrecondition" status code

    Examples:
      | signed-in user                    | timesheet-status |
      | staff granted role school admin   | APPROVED         |
      | staff granted role teacher        | APPROVED         |
      | staff granted role school admin   | DRAFT            |
      | staff granted role teacher        | DRAFT            |
      | staff granted role school admin   | CONFIRMED        |
      | staff granted role teacher        | CONFIRMED        |

  Scenario Outline: Staff cancel submit a timesheet for other staff
    Given an existing "SUBMITTED" timesheet for other staff "<other-staff-group>"
    When "<signed-in user>" signin system
    And user cancel submits the timesheet for other staff
    Then returns "<resp status-code>" status code
    And timesheet status changed to draft "<submit-status>"

    Examples:
      | signed-in user                    | resp status-code | submit-status   | other-staff-group                    |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |
      | staff granted role school admin   | OK               | successfully    | staff granted role school admin      |
      | staff granted role teacher        | PermissionDenied | unsuccessfully  | staff granted role teacher           |
