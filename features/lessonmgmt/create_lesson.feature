@runsequence
Feature: User creates lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario: School admin can create a lesson with all required fields
    Given user signed in as school admin
    When user creates a new lesson with all required fields
    Then returns "OK" status code
    And the lesson was created
    And Bob must push msg "CreateLessons" subject "Lesson.Created" to nats
    And student and teacher name must be updated correctly

  Scenario: School admin can create a lesson with "group teaching method" with all required fields
    Given user signed in as school admin
    And a class with id prefix "<prefix-class-id>" and a course with id prefix "<prefix-course-id>"
    When user creates a new lesson with "group" teaching method and all required fields
    Then returns "OK" status code
    And the lesson was created
    And student and teacher name must be updated correctly
      Examples:
      | prefix-class-id    | prefix-course-id    |
      | bdd-test-class-id- | bdd-test-course-id- |
