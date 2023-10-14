Feature: Modify virtual classroom room session time

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
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: teacher modify live lesson session time
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    When user modifies the session time in the live lesson
    Then returns "OK" status code
    And user gets session time in the live lesson