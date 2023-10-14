@quarantined
Feature: List study plan by course

    Scenario: List study plan by course
    Given a course and some study plans
    And a signed in "teacher" 
    And teacher archives some study plans
    When the teacher list study plan by course
    Then our system have to return list study plan by course correctly

   