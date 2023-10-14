# quarantined is for create multiple order with the same course, duplicate time range
@quarantined
Feature: Create order One Time Package

  Scenario: Create order slot base package success
    Given prepare data for create order slot base package
    When "school admin" submit order
    Then order slot base package is created successfully
    And receives "OK" status code
    And package upserted to student package table
    #And an event must be published to signal student course sync
