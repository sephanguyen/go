Feature: Create order One Time Material

  Scenario: Create bulk order success
    Given prepare data for create bulk order
    When "school admin" submit bulk order
    Then bulk order is created successfully
    And receives "OK" status code