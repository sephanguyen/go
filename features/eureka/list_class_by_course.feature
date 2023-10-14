Feature: List class by course feature

  Background: valid course background
    Given a valid course background
    And classes belong to course in bob

  Scenario: List class by course
    Given a signed in "teacher"
    When user list class by course
    Then returns "OK" status code
    And eureka must return correct list of class ids

  Scenario: List class by course and locations
    Given a signed in "teacher"
    When user list class by course and locations
    Then returns "OK" status code
    And eureka must return correct list of class ids

  Scenario: List class by course and not exist locations
    Given a signed in "teacher"
    When user list class by course and not exist locations
    Then returns "OK" status code
    And eureka must return nil list of class ids

  Scenario: List class by course and locations
    Given a signed in "teacher"
    And delete all classes belong to course
    When user list class by course and locations
    Then returns "OK" status code
    And eureka must return nil list of class ids
