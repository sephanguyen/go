Feature: Insert auto create timesheet flag log

  Scenario Outline: User upsert auto create flag and one logs record inserted
    Given "<signed-in user>" signin system
    And new flag data with "<flag-on>" status
    And count number "<flag-on>" status log record of user
    When user update a auto create timesheet flag
    Then returns "<resp status-code>" status code
    And a log record is inserted with status is "<flag-on>"

    Examples:
      | signed-in user                    | resp status-code | flag-on  |
      | staff granted role school admin   | OK               | false    |
