@quarantined
Feature: Import billing ratio

  Scenario Outline: Import valid csv file
    Given a billing ratio valid request payload with "<row condition>"
    When "<signed-in user>" importing billing ratio
    Then the valid billing ratio lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given a billing ratio valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing billing ratio
    Then the import billing ratio transaction is rolled back
    And the invalid billing ratio lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given a billing ratio invalid "<invalid format>" request payload
    When "<signed-in user>" importing billing ratio
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                         |
      | school admin   | no data                                                |
      | school admin   | header only                                            |
      | school admin   | number of column is not equal 7                        |
      | school admin   | mismatched number of fields in header and content      |
      | school admin   | wrong billing_ratio_id column name in header           |
      | school admin   | wrong start_date column name in header                 |
      | school admin   | wrong end_date column name in header                   |
      | school admin   | wrong billing_schedule_period_id column name in header |
      | school admin   | wrong billing_ratio_numerator column name in header    |
      | school admin   | wrong billing_ratio_denominator column name in header  |
      | school admin   | wrong is_archived column name in header                |
