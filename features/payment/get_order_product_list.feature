# quarantined is for create multiple order with the same course, duplicate time range
@quarantined
Feature: Get Order Product List In Student Billing

  Scenario: Get order product list success
    Given create data for order "<type of order>" "<type of product>"
    And "school admin" create "<type of order>" orders data for get list order product successfully
    When "school admin" get list order product with "<type of location>"
    Then check response data of "<type of order>" "<type of product>" with "<type of location>" successfully
  Examples:
    | type of product              | type of order | type of location |
    | one time package             | new           | valid location   |
    | one time material            | new           | valid location   |
    | one time fee                 | new           | valid location   |
    | recurring material           | new           | valid location   |
    | one time package             | new           | invalid location |
    | one time material            | new           | invalid location |
    | one time fee                 | new           | invalid location |
    | recurring material           | new           | invalid location |
    | one time package             | new           | empty location   |
    | one time material            | new           | empty location   |
    | one time fee                 | new           | empty location   |
    | recurring material           | new           | empty location   |
    # | recurring material           | loa           | valid location   |
    # | recurring material           | resume        | valid location   |
