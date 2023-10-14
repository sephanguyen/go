Feature: Get Product List Of Order In Order Detail Page

  Scenario Outline: Get product list of order success
    Given "school admin" create "<type of order>" order with "<type of product>" products successfully
    When "school admin" get product list of "<type of order>" order with "<type of filter>" filter
    Then get product list of order with "<type of response>" response successfully

    Examples:
    | type of product        | type of order      | type of filter     | type of response  |
    | one time package       | new                | valid              | non-empty         |
    | one time fee           | new                | valid              | non-empty         |
    | one time material      | new                | valid              | non-empty         |
    | one time package       | update             | valid              | non-empty         |
    | recurring package      | update             | valid              | non-empty         |
    | recurring material     | update             | valid              | non-empty         |
    | recurring material     | withdrawal         | valid              | non-empty         |
    | recurring fee          | withdrawal         | valid              | non-empty         |
    | recurring package      | withdrawal         | valid              | non-empty         |
    | recurring material     | graduate           | valid              | non-empty         |
    | recurring fee          | graduate           | valid              | non-empty         |
    | recurring package      | graduate           | valid              | non-empty         |
