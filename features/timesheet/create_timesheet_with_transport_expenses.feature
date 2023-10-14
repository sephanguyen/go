Feature: Create Timesheet with transportation expenses

Background:
    Given have timesheet configuration is on
    
Scenario Outline: Staff create timesheet with transportation expenses for themselves
    Given "<signed-in user>" signin system
    And new timesheet with transportation expenses
    When user creates a new timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code |
      | staff granted role school admin   | OK               |
      | staff granted role teacher        | OK               |

Scenario Outline: Staff create timesheet for current staff with invalid transportation data request
    Given "staff granted role school admin" signin system
    And new transportation data with "<invalid transportation argument>" for current staff
    When user creates a new timesheet
    Then returns "<resp status-code>" status code for "<invalid transportation argument>"

    Examples:
        | invalid transportation argument                       | resp status-code  |
        | transportation expenses list over 10                  | InvalidArgument   |
        | transportation type invalid                           | InvalidArgument   |
        | transportation expenses from null                     | InvalidArgument   |
        | transportation expenses from > 100 character          | InvalidArgument   |
        | transportation expenses to null                       | InvalidArgument   |
        | transportation expenses to > 100 character            | InvalidArgument   |
        | cost amount is empty                                  | InvalidArgument   |
        | cost amount is smaller than 0                         | InvalidArgument   |
        | round trip is empty                                   | InvalidArgument   |
        | transportation expenses remarks > 100 character       | InvalidArgument   |
