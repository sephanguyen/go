Feature: Import package course

  Scenario Outline: Import valid package course csv file
    Given an package course valid request payload with "<row condition>"
    When "<signed-in user>" importing package course
    Then the valid package course lines are imported successfully
    And the invalid package course lines are returned with error
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid package course csv file
    Given an package course invalid "<invalid format>" request payload
    When "<signed-in user>" importing package course
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 5                      |
      | school admin   | wrong package_id column name in header               |