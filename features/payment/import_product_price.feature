Feature: Import product price

  Scenario Outline: Import valid product price csv file with correct data
    Given an product price valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product price
    Then the valid product price lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
  
  Scenario Outline: Rollback failed import valid product price csv file with incorrect data
    Given an product price valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product price
    Then the import product price transaction is rolled back
    And the invalid product price lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |
  
  Scenario Outline: Import invalid product price csv file
    Given an product price invalid request payload with "<invalid format>"
    When "<signed-in user>" importing product price
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                         |
      | school admin   | no data                                                |
      | school admin   | header only                                            |
      | school admin   | number of column is not equal 5                        |
      | school admin   | mismatched number of fields in header and content      |
      | school admin   | wrong product_id column name in header                 |
      | school admin   | wrong billing_schedule_period_id column name in header |
      | school admin   | wrong quantity column name in header                   |
      | school admin   | wrong price column name in header                      |
      | school admin   | wrong price type column name in header                 |
      | school admin   | missing default_price value by product id              |
