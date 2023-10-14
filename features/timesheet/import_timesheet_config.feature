Feature: Import timesheet config

  Scenario Outline: Import valid csv file
    Given a timesheet config valid request payload
    When "staff granted role school admin" importing timesheet config
    Then the valid timesheet config lines are imported successfully
    And returns "OK" status code

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given a timesheet config valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing timesheet config
    Then the import timesheet config transaction is rolled back
    And the invalid timesheet config lines are returned with error
    Then returns "OK" status code for "<row condition>"

    Examples:
      | signed-in user                    | row condition          |
      | staff granted role school admin   | empty value row        |
      | staff granted role school admin   | invalid value row      |
      | staff granted role school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given a timesheet config invalid "<invalid format>" request payload
    When "<signed-in user>" importing timesheet config
    Then returns "InvalidArgument" status code for "<invalid format>"

    Examples:
      | signed-in user                    | invalid format                                     |
      | staff granted role school admin   | no data                                            |
      | staff granted role school admin   | header only                                        |
      | staff granted role school admin   | number of column is not equal 4                    |
      | staff granted role school admin   | mismatched number of fields in header and content  |
      | staff granted role school admin   | wrong timesheet_config_id column name in header    |
      | staff granted role school admin   | wrong config_type column name in header            |
      | staff granted role school admin   | wrong config_value column name in header           |
      | staff granted role school admin   | wrong is_archived column name in header            |