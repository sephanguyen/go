@blocker
Feature: Upsert books

  Background:
    Given a signed in "student"
    And fake a book content
    And seed faked book content

  Scenario: Get valid book content
    When user gets a "existing" book content
    Then returns "OK" status code
    And returns valid book content

  Scenario: Get valid book content
    When user gets a "not-existing" book content
    Then returns "NotFound" status code
