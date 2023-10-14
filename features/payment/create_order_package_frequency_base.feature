Feature: Create order frequency-base package

  Scenario Outline: Create order frequency-base package success
    Given prepare data for create order frequency-base package
    When "school admin" submit order
    Then order frequency-base package is created successfully
    And receives "OK" status code