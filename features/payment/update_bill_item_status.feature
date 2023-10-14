Feature: Update Bill Item Status

  Scenario Outline: Update bill items status with valid payload
    Given there is an existing bill items from order records with "<product>"
    And "<valid status>" request payload to update bill items status
    When "school admin" submitted the request using "<service>"
    And bill items status updated "successfully"
    And invoiced bill items will have invoiced order
    And response has no errors
    And receives "OK" status code

    Examples:
      | product                           | valid status | service  |
      | material type one time            | invoiced     | internal |
      | material type one time            | pending      | internal |
      | material type one time            | billed       | internal |
      | order with discount and prorating | invoiced     | internal |

#  Scenario Outline: Update bill items status with invalid payload
#    Given there is an existing bill items from order records with "material type one time"
#    And "invalid" request payload to update bill items status
#    When "school admin" submitted the request using "<service>"
#    Then bill items status updated "unsuccessfully"
#    And receives "InvalidArgument" status code
#
#    Examples:
#      | service  |
#      | internal |
