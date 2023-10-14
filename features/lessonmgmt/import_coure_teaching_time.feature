Feature: Import Course Teaching Time
  Background:
    Given user signed in as school admin 
    And have some centers
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And register some course's teaching time

  Scenario Outline: Import valid csv file
    Given user signed in as school admin
    And a valid course teaching time payload
    When importing course teaching time
    Then returns "OK" status code
