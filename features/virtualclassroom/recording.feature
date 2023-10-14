Feature: Recording Live Lesson

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

  Scenario: teacher start to recording
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "OK" status code
    And start recording state is updated
    When user stop recording
    Then returns "OK" status code
    And stop recording state is updated
    And recorded videos are saved
    When user get recorded videos on BO lesson detail
    Then returns "OK" status code
    And must return a list recorded video
    And user download each recorded videos
  
  Scenario: teacher start to recording in the live room from a lesson
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user starts recording in the live room
    Then returns "OK" status code
    And user gets the live room state recording has "started"
    When user stops recording in the live room
    Then returns "OK" status code
    And user gets the live room state recording has "stopped"
    And recorded videos are saved
    When user get recorded videos on BO lesson detail
    Then returns "OK" status code
    And must return a list recorded video
    And user download each recorded videos

  Scenario: teacher start to recording in the live room only
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user starts recording in the live room only
    Then returns "OK" status code
    And user gets the live room state recording has "started"
    When user stops recording in the live room only
    Then returns "OK" status code
    And user gets the live room state recording has "stopped"
    And recorded videos are saved in the live room

  Scenario: teacher start to recording the live lesson which already recorded
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    And user start to recording
    And returns "OK" status code
    And start recording state is updated
    When user start to recording
    And returns "AlreadyExists" status code

  Scenario: teacher start to recording the live room which already recorded
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    And user starts recording in the live room
    And returns "OK" status code
    And user gets the live room state recording has "started"
    When user starts recording in the live room
    And returns "AlreadyExists" status code

  Scenario: teacher start to recording then tries to stop recording two times
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "OK" status code
    And start recording state is updated
    When user stop recording
    Then returns "OK" status code
    And stop recording state is updated
    When user stop recording
    Then returns "NotFound" status code

  Scenario: teacher start to recording then tries to stop recording two times in the live room only
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user starts recording in the live room only
    Then returns "OK" status code
    And user gets the live room state recording has "started"
    When user stops recording in the live room only
    Then returns "OK" status code
    And user gets the live room state recording has "stopped"
    When user stops recording in the live room only
    Then returns "NotFound" status code

  Scenario: student can not start to recording
    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "PermissionDenied" status code

  Scenario: teacher get recording status
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start to recording
    Then returns "OK" status code
    And user get start recording state
    When user stop recording
    And user get stop recording state
