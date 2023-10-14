# quarantined is for create multiple order with the same course, duplicate time range
@quarantined
Feature: Create order multiple products

  Scenario: Create order multiple one time products success
    Given prepare data for creating order multiple one time products
    When "school admin" submit order
    Then order of multiple one time products is created successfully
    And receives "OK" status code


  Scenario: Create order multiple recurring products success
    Given prepare data for creating order multiple recurring products
    When "school admin" submit order
    Then order of multiple recurring products is created successfully
    And receives "OK" status code

  Scenario: Create order multiple one time and recurring products success
    Given prepare data for creating order multiple one time and recurring products
    When "school admin" submit order
    Then order of multiple one time and recurring products is created successfully
    And receives "OK" status code
