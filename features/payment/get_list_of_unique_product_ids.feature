Feature: Get Unique Product List

  Scenario: Get unque product list success
    Given create data for order "<type of order>" "<type of product>" for unique product
    And "school admin" create "<type of order>" orders data for get unique product successfully
    When "school admin" get unique product 
    Then check unique product of "<type of product>"
  Examples:
    | type of product              | type of order    |
    | one time material            | new              |
    | recurring material           | new              |