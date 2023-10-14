Feature: Get User Information

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: Teacher gets user information
    Given "teacher" signin system
    When user gets user information
    Then returns "OK" status code
    And user receives expected user information