@quarantined
Feature: Import product grade

  Scenario Outline: Import valid product grade csv file with correct data
    Given a product grade valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product grade
    Then the valid product grade lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | overwrite existing     |

  Scenario Outline: Rollback failed import valid product grade csv file with incorrect data
    Given a product grade valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product grade
    Then the import product grade transaction is rolled back
    And the invalid product grade lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid product grade csv file
    Given a product grade invalid request payload with "<invalid format>"
    When "<signed-in user>" importing product grade
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 2                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong product_id column name in header            |
      | school admin   | wrong grade_id column name in header              |
