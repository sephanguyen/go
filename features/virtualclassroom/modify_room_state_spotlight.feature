Feature: Modify Room State Spotlight

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

    Scenario: Teacher modify room state spotlight then unspotlight
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "adds" a spotlight user
    Then returns "OK" status code
    And user gets correct spotlight state
    When user "removes" a spotlight user
    Then returns "OK" status code
    And user gets correct spotlight state

    Scenario: Staff modifies room state spotlight then unspotlight
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user "adds" a spotlight user
    Then returns "OK" status code
    And user gets correct spotlight state
    When user "removes" a spotlight user
    Then returns "OK" status code
    And user gets correct spotlight state