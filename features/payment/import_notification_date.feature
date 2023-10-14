Feature: Import notification date
  @quarantined
  Scenario Outline: Import valid csv file
    Given a notification date valid request payload with "<row condition>"
    When "<signed-in user>" importing notification date
    Then the valid notification date lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given a notification date valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing notification date
    Then the import notification date transaction is rolled back
    And the invalid notification date lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given a notification date invalid "<invalid format>" request payload
    When "<signed-in user>" importing notification date
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 4                   |
      | school admin   | wrong notification_date_id column name in header  |
      | school admin   | wrong order_type column name in header            |
      | school admin   | wrong notification_date column name in header     |
      | school admin   | wrong is_archived column name in header           |
