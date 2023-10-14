Feature: Get Private Conversation IDs

   Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts with "AgoraTest TeacherLocal" first name and "AgoraTest" last name
    And have "6" student accounts with "StudentLocal AgoraTest" first name and "AgoraTest" last name
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing a virtual classroom session
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario Outline: teacher gets private conversation IDs
    Given an existing "teacher" user signin system
    And user already has existing private conversation with one of the student accounts
    When user gets private conversation IDs
    Then returns "OK" status code
    And user gets non-empty private conversation IDs
    
    When user gets private conversation IDs again
    Then returns "OK" status code
    And user gets the expected private conversation IDs
    
    # integration with GetConversationID, should return consistent result
    When an existing "student" user signin system
    And user gets private conversation ID with a "teacher" user
    Then returns "OK" status code
    And user gets the expected one private conversation ID