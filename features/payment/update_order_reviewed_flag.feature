Feature: Update Order Reviewed Flag

  # Scenario Outline: Update Order Reviewed Flag with valid payload
  #   Given an existing order from order records
  #   And "<valid status>" request payload to update order reviewed flag
  #   When "school admin" submitted the update order reviewed flag request
  #   And order reviewed flag updated "successfully"
  #   And update order reviewed flag response success
  #   And receives "OK" status code

  #   Examples:
  #     | valid status |
  #     | true         |
  #     | false        |

  Scenario Outline: Update Order Reviewed Flag with out of version request
    Given an existing order from order records
    And "true" request out of version payload to update order reviewed flag
    When "school admin" submitted the update order reviewed flag request
    And order reviewed flag updated "unsuccessfully"
    And update order reviewed flag response unsuccess
    And receives "FailedPrecondition" status code
