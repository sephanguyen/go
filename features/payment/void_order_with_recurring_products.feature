Feature: Void An Order With Recurring Products

  Scenario Outline: Void an order with recurring products successfully
    Given "school admin" create an "<type of order>" order with recurring products successfully
    When void an order with recurring products
    Then void an order with recurring products successfully
    And receives "OK" status code

    Examples:
      | type of order |
      | new           |
      | update        |
      | withdraw      |

  Scenario Outline: Void an order with recurring products successfully
    Given "school admin" create an "new" order with recurring products successfully
    When void an order with recurring products out of version
    Then void an order with recurring products unsuccessfully
    And receives "FailedPrecondition" status code
