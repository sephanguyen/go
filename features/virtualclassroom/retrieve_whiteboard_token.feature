Feature: Retrieve Whiteboard Token

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
    And lesson does not have existing room ID
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

    Scenario: student retrieve whiteboard token
    Given user signed as student who belong to lesson
    When user retrieves whiteboard token
    Then returns "OK" status code
    And user receives room ID and whiteboard token

    Scenario: teacher retrieve whiteboard token
    Given "teacher" signin system
    When user retrieves whiteboard token
    Then returns "OK" status code
    And user receives room ID and whiteboard token