Feature: Import fee period

  Scenario Outline: Import valid csv file
    Given an fee valid request payload with "<row condition>"
    When "<signed-in user>" importing fee
    Then the valid fee lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an fee valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing fee
    Then the import fee transaction is rolled back
    And the invalid fee lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |


  Scenario Outline: Import invalid csv file
    Given an fee invalid "<invalid format>" request payload
    When "<signed-in user>" importing fee
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 12                  |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong fee_id column name in header                |
