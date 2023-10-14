Feature: Generate billing items

  Scenario Outline: cron job will call this api GenerateBillingItems
    Given prepare data for scheduled generation of bill items recurring material with "<valid data>"
    When order is created and next upcoming billing date is within 30 days
    Then next billing items are generated
    And receives "OK" status code

    Examples:
      | valid data                                                                      |
      | order with single billed at order item                                          |
      | order with single billed at order item unique material                          |
