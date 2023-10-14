Feature: Update order Recurring Material

  Scenario Outline: Update order recurring material success
    Given prepare data for create order recurring material
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for update order recurring material
    When "school admin" submit order
    Then receives "OK" status code
    And update recurring material success

  Scenario Outline: Cancel order recurring material success
    Given prepare data for create order recurring material
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for cancel order recurring material
    When "school admin" submit order
    Then receives "OK" status code