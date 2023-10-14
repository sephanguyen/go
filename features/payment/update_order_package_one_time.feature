Feature: Create order One Time Package

#  Scenario: Update order one time package success
#    Given prepare data for create order one time package
#    And "school admin" submit order
#    And receives "OK" status code
#    And update data for update order one time package
#    When "school admin" submit order
#    Then order one time package is updated successfully
#    And receives "OK" status code
    #And an event must be published to signal student course sync

  Scenario: Cancel order one time package success
    Given prepare data for create order one time package
    And "school admin" submit order
    And receives "OK" status code
    And update data for cancel order one time package
    When "school admin" submit order
    Then receives "OK" status code
    #And an event must be published to signal student course sync
