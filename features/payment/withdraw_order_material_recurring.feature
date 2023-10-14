Feature: Withdraw order Recurring Material

  Scenario Outline: Withdraw order recurring material success
    Given prepare data for withdraw order recurring material with "<valid data>"
    When "school admin" submit order
    Then withdraw order recurring material is created successfully
    And receives "OK" status code

    Examples:
      | valid data                                           |
      | valid withdrawal request with disabled prorating     |
      | valid graduate request with disabled prorating       |
      | empty billed at order disabled prorating             |
      | empty upcoming billing disabled prorating            |
      | empty billed at order with prorating and discount    |
      | empty upcoming billing with prorating and discount   |
      | empty billing items                                  |
