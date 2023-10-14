Feature: Update order One Time Material

  Scenario: Update order one time material with bill_status invoiced success
    Given prepare data for update order one time material with bill_status invoiced
    When "school admin" submit order
    Then order one time material with bill_status invoiced is updated successfully
    And receives "OK" status code

  Scenario: Cancel order one time material with bill_status ordered success
    Given prepare data for cancel order one time material with bill_status ordered
    When "school admin" submit order
    Then order one time material with bill_status ordered is cancelled successfully
    And receives "OK" status code
