Feature: Import billing schedule

  Scenario Outline: Import valid csv file
    Given an billing schedule valid request payload with "<row condition>"
    When "<signed-in user>" importing billing schedule
    Then the valid billing schedule lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an billing schedule valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing billing schedule
    Then the import billing schedule transaction is rolled back
    And the invalid billing schedule lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |


  Scenario Outline: Import invalid csv file
    Given an billing schedule invalid "<invalid format>" request payload
    When "<signed-in user>" importing billing schedule
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 4                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong billing_schedule_id column name in header   |
      | school admin   | wrong name column name in header                  |
      | school admin   | wrong remarks column name in header               |
      | school admin   | wrong is_archived column name in header           |
