Feature: Modify Live Room Share Material State

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some medias

  Scenario: teacher try to share different materials with annotation updates
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user "enables" learners annotation in the live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room

    When user share a material with type is "audio" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "audio"

    Given student who is part of the live room
    When user joins an existing live room
    Then returns "OK" status code
    And user gets current material state of the live room is "audio"
    And user gets the expected annotation state in the live room is "enabled"

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user share a material with type is "pdf" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "pdf"

    When user share a material with type is "video" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "video"
    And user gets the expected annotation state in the live room is "enabled"
    
    Given student who is part of the live room
    When user joins an existing live room
    Then returns "OK" status code
    And user gets current material state of the live room is "video"
    And user gets the expected annotation state in the live room is "enabled"

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user stop sharing material in the live room
    Then returns "OK" status code
    And user gets empty current material state in the live room
    And user gets the expected annotation state in the live room is "enabled"

    When user disables all annotation in the live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room is "disabled"
    And have an uncompleted live room log with "4" joined attendees, "12" times getting room state, "6" times updating room state and "0" times reconnection

   Scenario: learner try to share and stop share material
    Given student who is part of the live room
    And user joins a new live room
    And returns "OK" status code
    And "student" receives channel and room ID and other tokens
    And user gets the expected annotation state in the live room is "enabled"

    When user share a material with type is "video" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "video"
    And user gets the expected annotation state in the live room is "enabled"

    When user share a material with type is "pdf" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "pdf"

    When user share a material with type is "audio" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "audio"

    When user stop sharing material in the live room
    Then returns "OK" status code
    And user gets empty current material state in the live room
    And user gets the expected annotation state in the live room is "enabled"
    And have an uncompleted live room log with "1" joined attendees, "7" times getting room state, "4" times updating room state and "0" times reconnection