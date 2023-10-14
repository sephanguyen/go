Feature: modify Room State Zoom Whiteboard

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

  Scenario: Teacher modify Room State Zoom Whiteboard
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user zoom whiteboard at in a virtual classroom session
    Then returns "OK" status code
    And user get zoom whiteboard state

  Scenario: Teacher modify Room State Zoom Whiteboard has the default value
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    And user get zoom whiteboard state with the default value

  Scenario: Teacher modify Room State Zoom Whiteboard has the default value with polling state
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start polling in a virtual classroom session
    Then returns "OK" status code
    And user get zoom whiteboard state with the default value

  Scenario: staff modifies Room State Zoom Whiteboard
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user zoom whiteboard at in a virtual classroom session
    Then returns "OK" status code
    And user get zoom whiteboard state