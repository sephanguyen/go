Feature: Update order Recurring Package

  Scenario Outline: Update order recurring package success
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for update order recurring package
    When "school admin" submit order
    Then receives "OK" status code
    And update recurring package success
    #And an event must be published to signal student course sync

  Scenario Outline: Cancel order recurring package success
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for cancel order recurring package
    When "school admin" submit order
    Then receives "OK" status code
    And update recurring package success
    #And an event must be published to signal student course sync

  Scenario Outline: Update order recurring package out of version
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for update order recurring package out of version
    When "school admin" submit order
    Then receives "FailedPrecondition" status code
    And update recurring package unsuccess with out version
