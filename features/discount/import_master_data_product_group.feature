Feature: Import product group

  Scenario Outline: Import valid product group csv file with correct data
    Given a product group valid request payload with correct "<row-condition>" data
    When "<signed-in user>" importing product group
    Then the valid product group lines with "<row-condition>" data are imported successfully
    And receives "OK" status code

    Examples:
      | signed-in user | row-condition          |
      | school admin   | all valid rows         |
      | hq staff       | overwrite existing     |

  Scenario Outline: Rollback failed import valid product group csv file with incorrect data
    Given a product group valid request payload with incorrect "<row-condition>" data
    When "<signed-in user>" importing product group
    Then the import product group transaction is rolled back

    Examples:
      | signed-in user | row-condition          |
      | school admin   | empty value row        |
      | school admin   | invalid value row      |
      | school admin   | valid and invalid rows |

  Scenario Outline: Import invalid product group csv file
    Given a product group invalid "<invalid format>" request payload
    When "<signed-in user>" importing product group
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invalid format                                    |
      | school admin   | no data                                           |
      | hq staff       | header only                                       |
      | school admin   | number of column is not equal 5                   |
      | hq staff       | mismatched number of fields in header and content |
      | school admin   | wrong product_group_id column name in header      |
      | hq staff       | wrong group_name column name in header            |
      | school admin   | wrong group_tag column name in header             |
      | hq staff       | wrong discount_type column name in header         |
      | school admin   | wrong is_archived column name in header           |
