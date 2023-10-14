@quarantine
Feature: Populate Stress Test Data for Local

  Background:
    Given user signed in as school admin for e2e
    When has a center for stress test
    And has a teacher for stress test
    And has a course for stress test
    And has a student for stress test
    And have some student subscriptions
    And have some medias

  Scenario: admin creates a set of lessons
    Given user signed in as school admin for e2e
    Then an existing set of "105" lessons