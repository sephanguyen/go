Feature: Get Billing Item in Order Details

  Scenario Outline: Get billing item in order details success
    Given "school admin" create "<type of order>" order with "<type of product>" products successfully
    When "school admin" get bill items of "<type of order>" order
    Then get bill items of order successfully

    Examples:
      | type of product                   | type of order      |
      | one time package                  | new                |
      | one time material                 | new                |
      | one time fee                      | new                |
      | recurring fee                     | new                |
      | recurring package                 | new                |
      | recurring material                | new                |
      | one time package                  | update             |
      | recurring package                 | update             |
      | recurring material                | update             |
      | custom billing                    | custom billing     |
      | recurring material                | withdrawal         |
      | recurring fee                     | withdrawal         |
      | recurring package                 | withdrawal         |
      | recurring material                | graduate           |
      | recurring fee                     | graduate           |
      | recurring package                 | graduate           |
