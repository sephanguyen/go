Feature: Get Conversation ID

   Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts with "AgoraTest TeacherLocal" first name and "AgoraTest" last name
    And have "12" student accounts with "StudentLocal AgoraTest" first name and "AgoraTest" last name
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing a virtual classroom session
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario Outline: teacher gets private conversation IDs with each students
    Given an existing "teacher" user signin system
    And user already has existing private conversations with other "2" "student" users
    When user gets private conversation ID with a "student" user
    Then returns "OK" status code
    And user gets non-empty conversation ID
    
    When user gets the same private conversation ID with a "student" user
    Then returns "OK" status code
    And user gets the expected private conversation ID
    
    When an existing "student" user signin system
    When user gets private conversation ID with a "teacher" user
    Then returns "OK" status code
    And user gets the expected private conversation ID

  Scenario Outline: teacher gets public conversation ID
    Given an existing "teacher" user signin system
    When user gets public conversation ID
    Then returns "OK" status code
    And user gets non-empty conversation ID
    
    When user gets the same public conversation ID
    Then returns "OK" status code
    And user gets the expected public conversation ID
    
    When an existing "student" user signin system
    And user gets public conversation ID
    Then returns "OK" status code
    And user gets the expected public conversation ID
