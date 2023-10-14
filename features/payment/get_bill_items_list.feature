# quarantined is for create multiple order with the same course, duplicate time range
@quarantined
Feature: Get Bill Items List

  Scenario Outline: Get bill items list success
    Given prepare data for get list bill items create "<type of bill>" "<type of product>"
    When "school admin" create "<type of bill>" orders data for get list bill items
    Then "school admin" get list bill items with "<filter>"

    Examples:
      | type of product       | type of bill | filter                     |
      | one time package      | new          | valid filter               |
      | one time material     | new          | valid filter               |
      | one time fee          | new          | valid filter               |
      | recurring material    | new          | valid filter               |
      | one time package      | new          | empty filter location      |
      | one time material     | new          | empty filter location      |
      | one time fee          | new          | empty filter location      |
      | recurring material    | new          | empty filter location      |
      | one time package      | new          | filter with empty response |
      | one time material     | new          | filter with empty response |
      | one time fee          | new          | filter with empty response |
      | recurring material    | new          | filter with empty response |
