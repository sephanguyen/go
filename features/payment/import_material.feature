Feature: Import material period

  Scenario Outline: Import valid csv file
    Given an material valid request payload with "<row condition>"
    When "<signed-in user>" importing material
    Then the valid material lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an material valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing material
    Then the import material transaction is rolled back
    And the invalid material lines are returned with error
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |


  Scenario Outline: Import invalid csv file
    Given an material invalid "<invalid format>" request payload
    When "<signed-in user>" importing material
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 13                  |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong material_id column name in header           |
