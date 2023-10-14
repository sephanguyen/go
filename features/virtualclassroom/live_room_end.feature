# @blocker
Feature: End Live Room

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some medias

  Scenario: Teacher end a live room after some state changes
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens

    When user starts recording in the live room only
    Then returns "OK" status code
    And user gets the live room state recording has "started"

    When user start polling with "3" options and "1" correct answers in the live room
    Then returns "OK" status code
    And user get current polling state in the live room has started

    When user "adds" a spotlighted user in the live room
    Then returns "OK" status code
    And user gets correct spotlight state in the live room

    When user zoom whiteboard in the live room
    Then returns "OK" status code
    And user gets whiteboard zoom state in the live room

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    
    When user "raise" hand in the live room
    Then returns "OK" status code
    And user get hands up state in the live room
    
    When user submit the answer "A,B" in the live room polling
    Then returns "OK" status code
    And user get polling answer state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user "enables" learners annotation in the live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room

    When user "disables" learners chat permission in the live room
    Then returns "OK" status code
    And user gets the expected chat permission state in the live room

    When user share a material with type is "pdf" in the live room
    Then returns "OK" status code
    And user gets current material state of the live room is "pdf"

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    And user gets current material state of the live room is "pdf"

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user end the live room
    Then returns "OK" status code
    And user get current polling state in the live room is empty
    And user gets empty current material state in the live room
    And user get all learner's hands up states to off in the live room
    And user gets the expected annotation state in the live room is "enabled"
    And user gets the expected chat permission in the live room is "enabled"
    And user gets empty spotlight in the live room
    And user gets whiteboard zoom state in the live room with default values
    And user gets the live room state recording has "stopped"
    And have a completed live room log with "4" joined attendees, "9" times getting room state, "8" times updating room state and "0" times reconnection