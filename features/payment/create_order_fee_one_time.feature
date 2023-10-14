Feature: Create order One Time Fee

  Scenario: Create order one time fee success
    Given prepare data for creating order one time fee
    When "school admin" submit order
    Then order one time fee is created successfully
    And receives "OK" status code