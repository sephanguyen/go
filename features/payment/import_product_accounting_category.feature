Feature: Import product accounting category
  @quarantined
  Scenario Outline: Import valid product accounting category csv file with correct data
    Given a product accounting category valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product accounting category
    Then the valid product accounting category lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | overwrite existing     |
  @quarantined
  Scenario Outline: Rollback failed import valid product accounting category csv file with incorrect data
    Given a product accounting category valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product accounting category
    Then the import product accounting category transaction is rolled back
    And the invalid product accounting category lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid product accounting category csv file
    Given a product accounting category invalid request payload with "<invalid format>"
    When "<signed-in user>" importing product accounting category
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                     |
      | school admin   | no data                                            |
      | school admin   | header only                                        |
      | school admin   | number of column is not equal 2 product_id only    |
      | school admin   | mismatched number of fields in header and content  |
      | school admin   | wrong product_id column name in csv header         |
      | school admin   | wrong accounting_category_id column name in header |