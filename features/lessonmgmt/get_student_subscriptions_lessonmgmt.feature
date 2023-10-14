Feature: Get student subscriptions

  Background:
    Given user signed in as school admin
    Given have some centers
    And have some teacher accounts
    And have some grades
    And have some student accounts
    And have some courses
    And have some student subscriptions

  Scenario: Admin get student subscriptions
    Given user signed in as school admin
    When user get list student subscriptions in lessonmgmt
    Then returns "OK" status code
    And got list student subscriptions in lessonmgmt
