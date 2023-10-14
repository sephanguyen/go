Feature: Get list of learners by lesson id

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing a virtual classroom session
    And students have enrollment status
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: teacher get list of learners in lesson
    Given "teacher" signin system
    When user gets list of learners in lesson
    Then returns "OK" status code
    And returns a list of students with enrollment status

  Scenario: student get list of learners in lesson
    Given user signed as student who belong to lesson
    When user gets list of learners in lesson
    Then returns "OK" status code
    And returns a list of students with enrollment status
