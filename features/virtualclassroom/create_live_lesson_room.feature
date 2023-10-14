Feature: Create Live Lesson Room After Creating Lesson

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: Teacher create a lesson and a room ID is automatically created
    Given "teacher" signin system
    When user creates a virtual classroom session
    Then returns "OK" status code
    And lesson has an existing room ID with wait