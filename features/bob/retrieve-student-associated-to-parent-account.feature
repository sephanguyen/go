@quarantined
Feature: Retrieve students associated to parent account

  Background:
    Given multiple students profile in DB
    Then returns "OK" status code

  Scenario: Parent have children that are manabie's students
    Given "staff granted role school admin" signin system
    When create handsome father as a parent and the relationship with his children who're students at manabie
    Then returns "OK" status code

  Scenario: Retrieve students associated to parent account
    Given "staff granted role school admin" signin system
    And create handsome father as a parent and the relationship with his children who're students at manabie
    And a signed in parent
    When retrieve students profiles associated to parent account
    And fetched students exactly associated to parent
    Then returns "OK" status code