Feature: Join a live lesson

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

  Scenario: Teacher and student join a virtual classroom session / live lesson
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    And "teacher" receives room ID and other tokens

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    And "student" receives room ID and other tokens

  Scenario: Student joins virtual classroom session first
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    And "student" receives room ID and other tokens
    And user gets learners chat permission to "enabled" with wait

  Scenario: Teacher modifies disabled chat and stay disabled after joining sesssion
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "enabled" with wait
    When user "disables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"
    When user join a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled" with wait

  Scenario: staff joins a virtual classroom session / live lesson
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    And "staff" receives room ID and other tokens
