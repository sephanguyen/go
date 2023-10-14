Feature: Import package

  Scenario Outline: Import valid package csv file
    Given an package valid request payload with "<row condition>"
    When "<signed-in user>" importing package
    Then the valid package lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Import valid package csv file
    Given an package valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing package
    Then the import package transaction is rolled back
    And the invalid package lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid package csv file
    Given an package invalid "<invalid format>" request payload
    When "<signed-in user>" importing package
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 15                     |
      | school admin   | mismatched number of fields in header and content    |
      | school admin   | wrong package_id column name in header               |