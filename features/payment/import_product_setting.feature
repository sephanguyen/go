Feature: Import product setting

  Scenario Outline: Import valid product setting csv file with correct data
    Given a product setting valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing product setting
    Then the valid product setting lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | overwrite existing     |

  Scenario Outline: Rollback failed import valid product setting csv file with incorrect data
    Given a product setting valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing product setting
    Then the import product setting transaction is rolled back
    And the invalid product setting lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid product setting csv file
    Given a product setting invalid request payload with "<invalid format>"
    When "<signed-in user>" importing product setting
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                             |
      | school admin   | no data                                                    |
      | school admin   | header only                                                |
      | school admin   | number of column is not equal 5                            |
      | school admin   | mismatched number of fields in header and content          |
      | school admin   | incorrect product_id column name in header                 |
      | school admin   | incorrect is_enrollment_required column name in header     |
