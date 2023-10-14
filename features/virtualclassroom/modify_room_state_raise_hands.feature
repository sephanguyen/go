Feature: Modify virtual classroom room raise hands state

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

  Scenario: student try to raise hand and after teacher will fold learner's hand
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user raise hand in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user fold a learner's hand in a virtual classroom session
    Then returns "OK" status code
    And user get all learner's hands up states who all have value is off

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user raise hand in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user fold hand all learner
    Then returns "OK" status code
    And user get all learner's hands up states who all have value is off
    And have a uncompleted log with "3" joined attendees, "4" times getting room state, "4" times updating room state and "0" times reconnection

  Scenario: student try to raise hand and after hand off
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user raise hand in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user get hands up state
    Then returns "OK" status code

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user hand off in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection

  Scenario: student try to raise hand and after staff will fold learner's hand
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user raise hand in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state

    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user fold a learner's hand in a virtual classroom session
    Then returns "OK" status code
    And user get all learner's hands up states who all have value is off