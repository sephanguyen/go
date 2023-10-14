@runsequence @quarantined
Feature: List students by lesson

  Scenario: teacher list students in lesson
    Given a lesson with some lesson members
    When teacher list students in that lesson
    Then returns a list of students

  Scenario: student list students in lesson
    Given a lesson with some lesson members
    When student list students in that lesson
    Then returns a list of students

  Scenario: student list students in lesson in case some students has been removed from lesson
    Given a lesson with some lesson members
      And some students has been removed from the lesson
    When student list students in that lesson
    Then our system have to returns a list of students correctly

  Scenario: student list students with given their name in the lesson
    Given a lesson with some lesson members given their name
    When student list students in that lesson
    Then returns a list of students