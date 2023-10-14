@quarantined
Feature: Get student subscriptions

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions

  Scenario: Admin get student subscriptions
    Given user signed in admin
    When user get list student subscriptions
    Then returns "OK" status code
    And got list student subscriptions