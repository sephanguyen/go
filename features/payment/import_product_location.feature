@wip @quarantine
Feature: Import product location period

  Scenario Outline: Import valid product location csv file with correct data
    Given an product location valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product location
    Then the valid product location lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid product location csv file with incorrect data
    Given an product location valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product location
    Then the import product location transaction is rolled back
    And the invalid product location lines are returned with error
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
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 2                      |
      | school admin   | wrong product_id column name in header               |