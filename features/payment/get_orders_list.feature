Feature: Get Orders List

  Scenario: Get order list Success
    Given prepare data for getting order list
    When "school admin" create orders data for getting order list
    Then "school admin" get order list after creating orders with "<filter>"
    
    Examples:
    | filter                                |
    | without filter                        |
    | empty filter                          |
    | filter with empty response            |
    | filter with paginated result          |
    | keyword filter no student match       |
    | keyword filter case insensitive match |
    | product id filter empty response      |
    | valid product id filter               |
    | only is not reviewed filter           |
