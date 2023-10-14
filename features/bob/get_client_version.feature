Feature: Get client version

  Scenario: unauthenticated user try to get client version
    Given an invalid authentication token
    When a user get client version
    Then returns "Unauthenticated" status code

  Scenario: authenticated user try to get client version
    Given a signed in student
    When a user get client version
    Then returns "OK" status code
    And bob must returns client version from config