Feature: Modify Live Room State Annotation

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

  Scenario: teacher try to enable and after disable annotation
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user "enables" learners annotation in the live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room

    Given student who is part of the live room
    When user joins an existing live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user "disables" learners annotation in the live room
    Then returns "OK" status code
    And user gets the expected annotation state in the live room
    And have an uncompleted live room log with "3" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection