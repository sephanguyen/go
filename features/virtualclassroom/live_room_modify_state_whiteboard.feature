Feature: Modify Live Room Whiteboard Zoom

  Scenario: teacher modifies the whiteboard zoom state
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user zoom whiteboard in the live room
    Then returns "OK" status code
    And user gets whiteboard zoom state in the live room

  Scenario: teacher gets the whiteboard zoom state in the default value
    Given "teacher" signin system
    When user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    And user gets whiteboard zoom state in the live room with default values