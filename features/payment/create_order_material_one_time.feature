Feature: Create order One Time Material

    Scenario Outline: Create order one time material success
        Given prepare data for create order one time material with "<discount_type>"
        When "school admin" submit order
        Then order one time material is created successfully
        And receives "OK" status code

        Examples:
            | discount_type      |
            | product discount   |
            | org-level discount |
