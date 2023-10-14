Feature: Get list of learners by lessons

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And existing a virtual classroom sessions
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: teacher get list of learners in lesson
    Given "teacher" signin system
    When user gets list of learners from lessons
    Then returns "OK" status code
    And returns a list of students

  Scenario: student get list of learners in lesson 
    Given user signed as student who belong to lesson
    When user gets list of learners from lessons
    Then returns "OK" status code
    And returns a list of students