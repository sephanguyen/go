Feature: update timesheet with transport expenses
  Background:
    Given have timesheet configuration is on
    
  Scenario Outline: School Admin update timesheet with invalid value
    Given "staff granted role school admin" signin system
    And new update "<invalid transportation argument>" for timesheet with transportation expenses data for current staff
    When user update a timesheet
    Then returns "<resp status-code>" status code for "<invalid argument>"

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
        | transportation expenses remarks > 100 character       | InvalidArgument   |

  Scenario Outline: School Admin Update timesheet with transport expenses actions create, update, delete
    Given "staff granted role school admin" signin system
    And new update timesheet with "<action>" transportation expenses data for current staff
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
      | have-10,insert-2,delete-1   | InvalidArgument  |
