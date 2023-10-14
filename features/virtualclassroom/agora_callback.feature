Feature: agora callback

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

  Scenario: Agora callback when stop recording
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    And user start to recording
    And returns "OK" status code
    And start recording state is updated
    And a request exit recording service
    And a valid Agora signature in its header
    When agora callback
    And returns "OK" status code
    And stop recording state is updated
