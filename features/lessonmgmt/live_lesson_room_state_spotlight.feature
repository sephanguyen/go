Feature: Modify live lesson room spotlight state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And an existing lesson

  Scenario: teacher try to spotlight for user
    Given user signed in as teacher
    When user enable spotlight for student 
    Then returns "OK" status code
    And user get spotlighted user

  Scenario: teacher try to unspotlight for user
    Given user signed in as teacher
    And user enable spotlight for student
    When user disable spotlight for student 
    Then returns "OK" status code
    And user get spotlighted user
