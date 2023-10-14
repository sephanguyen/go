Feature: Create order Recurring Material

  Scenario Outline: Create order recurring material success
    Given prepare data for create order recurring material with "<valid data>"
    When "school admin" submit order
    Then order recurring material is created successfully
    And receives "OK" status code

    Examples:
      | valid data                                                                      |
      | order with discount applied fix amount with null recurring valid duration       |
      | order with discount applied fix amount with finite recurring valid duration     |
      | order with discount applied percent type with null recurring valid duration     |
      | order with discount applied percent type with finite recurring valid duration   |
      | order with discount and prorating                                               |
      | order with empty billed at order items                                          |
      | order with single billed at order item                                          |
      | order with multiple billed at order item                                        |
      | order with prorating applied                                                    |
      | order without prorating applied                                                 |