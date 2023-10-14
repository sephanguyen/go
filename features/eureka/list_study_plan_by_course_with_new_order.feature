@quarantined
Feature: List student by course with new order feature

  Scenario: Add new study plan
    Given a valid "current teacher" token
    When Add new study plans
    Then returns "OK" status code
    And  All study plans were inserted

  Scenario: Filter study pland with new order collation
    And a signed in "teacher" 
    When user get list study plans and filter with new order collation
    Then returns "OK" status code