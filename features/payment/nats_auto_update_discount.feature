Feature: Discount automation

  Scenario Outline: Update discount of recurring product
    Given prepare data for create order recurring "<product>"
    And "school admin" submit order
    When student tagged for org level "<discount>" discount
    And discount service sends data for discount update
    Then recurring "<product>" "<discount>" with discount is updated successfully

    Examples:
      | product                      | discount                    |
      | material                     | single-parent               |
      | material                     | family                      |
      | material                     | employee full-time          |
      | material                     | employee part-time          |
      | fee                          | family                      |
      | fee                          | single-parent               |
      | fee                          | employee full-time          |
      | fee                          | employee part-time          |
      | frequency-base package       | single-parent               |
      | frequency-base package       | family                      |
      | frequency-base package       | employee full-time          |
      | frequency-base package       | employee part-time          |
      | schedule-base package        | single-parent               |
      | schedule-base package        | family                      |
      | schedule-base package        | employee full-time          |
      | schedule-base package        | employee part-time          |