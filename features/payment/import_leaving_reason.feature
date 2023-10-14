Feature: Import leaving reason

  Scenario Outline: Import valid csv file
    Given an leaving reason valid request payload with "<row condition>"
    When "<signed-in user>" importing leaving reason
    Then the valid leaving reason lines are imported successfully
    And the invalid leaving reason lines are returned with error
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |


  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an leaving reason valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing leaving reason
    Then the import leaving reason transaction is rolled back
    And the invalid leaving reason lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given an leaving reason invalid "<invalid format>" request payload
    When "<signed-in user>" importing leaving reason
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 5                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong leaving_reason_id column name in header     |
      | school admin   | wrong name column name in header                  |
      | school admin   | wrong remarks column name in header               |
      | school admin   | wrong is_archived column name in header           |
