Feature: Void An Order With One-time Products

  Scenario: Void an order with one-time products successfully
    Given "school admin" create "<type of order>" order with "<type of product>" products successfully
    When "school admin" void an "<type of order>" order with one-time-products
    Then void an order with one-time-products successfully
    And receives "OK" status code

    Examples:
      | type of product                   | type of order      |
      | one time package                  | new                |
      | one time material                 | new                |
      | one time fee                      | new                |
      | custom billing                    | custom billing     |

  Scenario: Void an order with one-time products successfully
      Given "school admin" create "<type of order>" order with "<type of product>" products successfully
      When "school admin" void an "<type of order>" order with out of version request
      Then void an order with one-time-products unsuccessfully
      And receives "FailedPrecondition" status code

    Examples:
      | type of product                   | type of order      |
      | one time package                  | new                |
      