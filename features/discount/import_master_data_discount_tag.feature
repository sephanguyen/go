Feature: Import discount tag

  Scenario Outline: Import valid csv file
    Given an discount tag valid request payload with "<row condition>"
    When "<signed-in user>" importing discount tag
    Then the valid discount tag lines are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | all valid rows         |

  Scenario Outline: Rollback failed import valid setting csv file with incorrect data
    Given an discount tag valid request payload with incorrect data with "<row condition>"
    When "<signed-in user>" importing discount tag
    Then the import discount tag transaction is rolled back
    And the invalid discount tag lines are returned with error
    Then receives "OK" status code

    Examples:
      | signed-in user | row condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |


  Scenario Outline: Import invalid csv file
    Given an discount tag invalid "<invalid format>" request payload
    When "<signed-in user>" importing discount tag
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | school admin   | header only                                       |
      | school admin   | number of column is not equal 4                   |
      | school admin   | mismatched number of fields in header and content |
      | school admin   | wrong discount_tag_id column name in header       |
      | school admin   | wrong name column name in header                  |
      | school admin   | wrong selectable column name in header            |
