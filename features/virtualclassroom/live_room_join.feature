Feature: Join a Live Room

  Scenario: Teacher joins a new live room and student joins the existing live room
    Given "teacher" signin system
    When user joins a new live room
    Then returns "OK" status code
    And "teacher" receives channel and room ID and other tokens

    Given "student" signin system
    When user joins an existing live room
    Then returns "OK" status code
    And "student" receives channel and room ID and other tokens
    And user gets the expected chat permission in the live room is "enabled" with wait
    And user gets the expected annotation state in the live room is "enabled" with wait
    And have an uncompleted live room log with "2" joined attendees, "2" times getting room state, "0" times updating room state and "0" times reconnection