@quarantined
Feature: Import package quantity type mapping

  Scenario Outline: Import valid package quantity type mapping csv file with correct data
    Given an package quantity type mapping valid request payload with correct data with "<row condition>"
    When "<signed-in user>" importing package quantity type mapping
    Then the valid package quantity type mapping lines are imported successfully
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | overwrite existing     |

  Scenario Outline: Rollback failed import valid package quantity type mapping csv file with incorrect data
    Given an package quantity type mapping valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing package quantity type mapping
    Then the import package quantity type mapping transaction is rolled back
    And the invalid package quantity type mapping lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid package quantity type mapping csv file
    Given a package quantity type mapping invalid request payload with "<invalid format>"
    When "<signed-in user>" importing package quantity type mapping
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 2                      |
      | school admin   | mismatched number of fields in header and content    |
      | school admin   | incorrect package_type column name in header         |
      | school admin   | incorrect quantity_type column name in header        |
