Feature: Import tax

  Scenario Outline: Import valid csv file
    Given a tax valid request payload with "<row condition>"
    When "<signed-in user>" importing tax
    Then the valid tax lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given a tax valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing tax
    Then the import tax transaction is rolled back
    And the invalid tax lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |


  Scenario Outline: Import invalid csv file
    Given a tax invalid "<invalid format>" request payload
    When "<signed-in user>" importing tax
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 6                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong tax_id column name in header                |
      | school admin   | wrong name column name in header                  |
      | school admin   | wrong tax_percentage column name in header        |
      | school admin   | wrong tax_category column name in header          |
      | school admin   | wrong default_flag column name in header          |
      | school admin   | wrong is_archived column name in header           |
