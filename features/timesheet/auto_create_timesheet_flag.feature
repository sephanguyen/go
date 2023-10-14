Feature: Upsert auto create timesheet flag

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: User upsert auto create timesheet flag
    Given "<signed-in user>" signin system
    And new flag data with "<flag-on>" status
    When user update a auto create timesheet flag
    Then returns "<resp status-code>" status code
    And flag status changed to "<flag-on>"

    Examples:
      | signed-in user                    | resp status-code | flag-on  |
      | staff granted role school admin   | OK               | false    |
  
  Scenario Outline: Invalid user upsert auto create timesheet flag
    Given "<signed-in user>" signin system
    And new flag data with "<flag-on>" status
    When user update a auto create timesheet flag
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code | flag-on  |
      | staff granted role school admin   | OK               | false    |
      | staff granted role teacher        | PermissionDenied | false    |
