Feature: Create order schedule-base package

  Scenario Outline: Create order schedule-base package success
    Given prepare data for create order schedule-base package
    When "school admin" submit order
    Then order schedule-base package is created successfully
    And receives "OK" status code
    And package upserted to student package table
    #And an event must be published to signal student course sync