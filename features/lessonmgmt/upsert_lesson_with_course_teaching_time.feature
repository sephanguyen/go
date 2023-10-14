@runsequence
Feature: User creates lesson

  Background:
    When enter a school
    Given have some centers
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And "enable" Unleash feature with feature name "Lesson_LessonManagement_CourseTeachingTime"

  Scenario: School admin can create a one-time lesson with course teaching time
    Given user signed in as school admin
    And register some course's teaching time
    When user creates a "draft" lesson with "teachers" in "lessonmgmt"
    Then returns "OK" status code
    And the lesson was created in lessonmgmt
    And the lesson have course's teaching time info
