Feature: Leave a Live Room

  Scenario: teacher can leave a live room
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user leaves the current live room
    Then returns "OK" status code
    And have an uncompleted live room log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

  Scenario: student can leave a live room
    Given "student" signin system
    And user joins a new live room
    And returns "OK" status code
    And "student" receives channel and room ID and other tokens
    When user leaves the current live room
    Then returns "OK" status code
    And have an uncompleted live room log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection