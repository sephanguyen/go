Feature: Get Live Room State

  Scenario: Teacher joins a new live room and gets live room state in default empty state
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user gets live room state
    Then returns "OK" status code
    And live room state is in default empty state

  Scenario: Student joins a new live room and gets live room state in default empty state
    Given "student" signin system
    And user joins a new live room
    And returns "OK" status code
    And "student" receives channel and room ID and other tokens
    When user gets live room state
    Then returns "OK" status code
    And live room state is in default empty state