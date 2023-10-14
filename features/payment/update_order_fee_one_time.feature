Feature: Update order One Time Fee

  Scenario: Update order one time fill with bill_status invoiced success
    Given prepare data for update order one time fee with bill_status invoiced
    When "school admin" submit order
    Then order one time fee with bill_status invoiced is updated successfully
    And receives "OK" status code

  Scenario: Cancel order one time fee with bill_status ordered success
    Given prepare data for cancel order one time fee with bill_status ordered
    When "school admin" submit order
    Then order one time fee with bill_status ordered is cancelled successfully
    And receives "OK" status code

  Scenario: Update order one time fill with bill_status ordered optimistic locking
    Given prepare data for update order one time fee of out version 
    When "school admin" submit order
    Then order one time fee with bill_status invoiced is not updated 
    And receives "FailedPrecondition" status code