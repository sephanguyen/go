Feature: Update Student Product Status

  Scenario Outline: Update student product status to cancelled when effective date of withdraw/graduate reached
    Given prepare data for "<order type>" order with valid effective date
    When "school admin" submit order
    And the scheduled job runs for "<order type>" on the effective date of order
    Then student product status changed to "<status>"

    Examples:
      | order type            | status                     |
      | withdrawal            | CANCELLED                  |
      | graduation            | CANCELLED                  |
      | LOA                   | PAUSED                     |
      | cronjob               | CANCELLED or PAUSED        |
