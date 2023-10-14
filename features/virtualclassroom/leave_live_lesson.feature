Feature: Leave a live lesson

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

  Scenario: Teacher can leave a virtual classroom session / live lesson
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user leaves the current virtual classroom session
    Then returns "OK" status code

  Scenario: Student can leave a virtual classroom session / live lesson
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user leaves the current virtual classroom session 
    Then returns "OK" status code

  Scenario: Teacher share a material and student can still leave live lesson
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    And returns "OK" status code
    And user get current material state of a virtual classroom session is pdf

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user get current material state of a virtual classroom session is pdf
    Then returns "OK" status code
    When user leaves the current virtual classroom session 
    Then returns "OK" status code 
