Feature: Create order One Time Package

  Scenario: Create order one time package success
    Given prepare data for create order one time package
    When "school admin" submit order
    Then order one time package is created successfully
    And receives "OK" status code
    And package upserted to student package table
    #And an event must be published to signal student course sync
