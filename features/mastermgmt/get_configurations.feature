@blocker
Feature: Get all configurations
  
  Background: Given condition
    Given configurations value existed on DB

  Scenario: Get all configurations
    Given "school admin" signin system
    When user gets configurations with "empty" keyword
    Then returns "OK" status code
    And configurations are returned all items

