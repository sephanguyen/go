Feature: Import product discount

  Scenario Outline: Import valid product discount csv file with correct data
    Given an product discount valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product discount
    Then the valid product discount lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | overwrite existing     |

  Scenario Outline: Rollback failed import valid product discount csv file with incorrect data
    Given an product discount valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product discount
    Then the import product discount transaction is rolled back
    And the invalid product discount lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid product discount csv file
    Given an product discount invalid request payload with "<invalid format>"
    When "<signed-in user>" importing product discount
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 2                      |
      | school admin   | mismatched number of fields in header and content    |
      | school admin   | wrong product_id column name in header               |
      | school admin   | wrong discount_id column name in header              |