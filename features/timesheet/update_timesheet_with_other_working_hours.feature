Feature: update timesheet with other working hours

  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: School Admin update timesheet with invalid value
    Given "staff granted role school admin" signin system
    And new update "<invalid argument>" for timesheet with other working hours data for current staff
    When user update a timesheet
    Then returns "<resp status-code>" status code for "<invalid argument>"

    Examples:
        | invalid argument                                  | resp status-code  |
        | empty timesheet id                                | InvalidArgument   |
        | remark > 500 characters                           | InvalidArgument   |
        | other working hours list over 5                   | OK                |
        | other working hours working type empty            | InvalidArgument   |
        | other working hours start time null               | InvalidArgument   |
        | other working hours end time null                 | InvalidArgument   |
        | other working hours end time before start time    | InvalidArgument   |
        | other working hours end time == start time        | InvalidArgument   |
        | other working hours remarks > 100 character       | InvalidArgument   |
        | other working hours working type invalid          | Internal          |

  Scenario Outline: School Admin Update timesheet with other working hours actions create, update, delete
    Given "staff granted role school admin" signin system
    And new update timesheet with "<action>" other working hours data for current staff
    When user update a timesheet
    Then returns "<resp status-code>" status code for "<action>"

    Examples:
      | action                      | resp status-code |
      | insert                      | OK               |
      | update                      | OK               |
      | delete                      | OK               |
      | insert,delete               | OK               |
      | insert,update               | OK               |
      | update,delete               | OK               |
      | insert,update,delete        | OK               |
      | have-5,insert-2,delete-1    | InvalidArgument  |
