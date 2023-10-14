Feature: Import associated products by material

  Scenario Outline: Import valid csv file
    Given associated products by material valid request payload with "<row condition>"
    When "<signed-in user>" importing associated products by material
    Then the valid associated products by material lines are imported successfully
    And the invalid associated products by material lines are returned with error
    And receives "OK" status code

    Examples:
      | signed-in user | row condition      |
      | school admin   | all valid rows     |
      | school admin   | overwrite existing |

  Scenario Outline: Rollback failed import valid associated products by material csv file with incorrect data
    Given associated products by material valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing associated products by material
    Then the import associated products by material transaction is rolled back
    And the invalid associated products by material lines are returned with error
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid csv file
    Given associated products by material invalid "<invalid format>" request payload
    When "<signed-in user>" importing associated products by material
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 2 package_id only   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong package_id column name in csv header        |
      | school admin   | wrong course_id column name in csv header         |
      | school admin   | wrong material_id column name in csv header       |
      | school admin   | wrong available_from column name in csv header    |
      | school admin   | wrong available_until column name in csv header   |
