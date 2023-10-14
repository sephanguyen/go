Feature: Import discount

  Scenario Outline: Import valid csv file
    Given an discount valid request payload with "<row condition>"
    When "<signed-in user>" importing discount
    Then the valid discount lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |


  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an discount valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing discount
    Then the import discount transaction is rolled back
    And the invalid discount lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given an discount invalid "<invalid format>" request payload
    When "<signed-in user>" importing discount
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                       |
      | school admin   | no data                                              |
      | school admin   | header only                                          |
      | school admin   | number of column is not equal 10                     |
      | school admin   | mismatched number of fields in header and content    |
      | school admin   | wrong discount_id column name in header              |
      | school admin   | wrong name column name in header                     |
      | school admin   | wrong discount_type column name in header            |
      | school admin   | wrong discount_amount_type column name in header     |
      | school admin   | wrong discount_amount_value column name in header    |
      | school admin   | wrong recurring_valid_duration column name in header |
      | school admin   | wrong available_from column name in header           |
      | school admin   | wrong available_until column name in header          |
      | school admin   | wrong remarks column name in header                  |
      | school admin   | wrong is_archived column name in header              |