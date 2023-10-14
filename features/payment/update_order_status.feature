Feature: Update Order Status

  Scenario Outline: Update order status with valid payload
    Given there is an existing order from order records
    And "<valid status>" request payload to update order status
    When "school admin" submitted the update order request
    And order status updated "successfully"
    And update order status response has no errors
    And receives "OK" status code

    Examples:
      | valid status |
      | invoiced     |
      | submitted    |

  Scenario Outline: Update order status with invalid payload
    Given there is an existing order from order records
    And "invalid" request payload to update order status
    When "school admin" submitted the update order request
    Then order status updated "unsuccessfully"
    And receives "InvalidArgument" status code