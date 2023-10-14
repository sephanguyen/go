Feature: Create order with enrollment required tag

  Scenario: Create order with enrollment required tag with specific student status
    Given prepare data for create order with enrollment required tag set to to "<tag>" and student status set to "<status>"
    When "school admin" submit order
    Then permission to order enrollment required tag set to "<tag>" and student status set to "<status>" is validated

    Examples:
      | tag       | status    |
      | true      | enrolled  |
      | true      | potential |
      | false     | enrolled  |
      | false     | potential |
