# @blocker
Feature: End a live lesson

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

  Scenario: Teacher end a live lesson after some state changes
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "OK" status code
    And start recording state is updated
    When user start polling with "2" options and "0" correct answers in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started
    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state
    When user "disables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is pdf

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user get current material state of a virtual classroom session is pdf
    Then returns "OK" status code
    When user raise hand in a virtual classroom session
    Then returns "OK" status code
    And user get hands up state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user end the live lesson
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session is empty
    And user get current material state of a virtual classroom session is empty
    And user get all learner's hands up states who all have value is off
    And all annotation state is disable
    And user gets learners chat permission to "enabled"
    And stop recording state is updated

  Scenario: Staff end a live lesson after some state changes
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "OK" status code
    And start recording state is updated
    When user start polling with "2" options and "0" correct answers in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started
    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state
    When user "disables" chat of learners in a virtual classroom session
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is pdf

    When user end the live lesson
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session is empty
    And user get current material state of a virtual classroom session is empty
    And user get all learner's hands up states who all have value is off
    And all annotation state is disable
    And user gets learners chat permission to "enabled"
    And stop recording state is updated