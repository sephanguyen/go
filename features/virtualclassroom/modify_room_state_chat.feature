Feature: Modify virtual classroom room chat state

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

  Scenario: teacher tries to enable learners chat after disabling
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "disables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "enables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "enabled"
    And have a uncompleted log with "2" joined attendees, "2" times getting room state, "2" times updating room state and "0" times reconnection

 Scenario: teacher tries to disable learners chat after enable
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "enables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "enabled"

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "disables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"
    And have a uncompleted log with "2" joined attendees, "2" times getting room state, "2" times updating room state and "0" times reconnection

  Scenario: staff tries to disable learners chat after enable
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "enables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "enabled"