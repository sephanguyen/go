Feature: Create Timesheet with other working hours

Background:
    Given have timesheet configuration is on

Scenario Outline: Staff create timesheet with other working hours and remark for themselves
    Given "<signed-in user>" signin system
    And new timesheet data with other working hours and remark for current staff
    When user creates a new timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code |
      | staff granted role school admin   | OK               |
      | staff granted role teacher        | OK               |

Scenario Outline: Staff create timesheet for current staff with invalid request
    Given "staff granted role school admin" signin system
    And new data with "<invalid argument>" for current staff
    When user creates a new timesheet
    Then returns "<resp status-code>" status code for "<invalid argument>"

    Examples:
        | invalid argument                                  | resp status-code  |
        | empty StaffId                                     | InvalidArgument   |
        | empty LocationId                                  | InvalidArgument   |
        | null Date                                         | InvalidArgument   |
        | remark > 500 characters                           | InvalidArgument   |
        | other working hours list over 5                   | InvalidArgument   |
        | other working hours working type empty            | InvalidArgument   |
        | other working hours start time null               | InvalidArgument   |
        | other working hours end time null                 | InvalidArgument   |
        | other working hours end time before start time    | InvalidArgument   |
        | other working hours remarks > 100 character       | InvalidArgument   |
        | other working hours working type invalid          | Internal          |
