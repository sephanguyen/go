Feature: Modify Live Room Hands Up

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

  Scenario: student try to raise hand and teacher will fold learner's hand after
    Given student who is part of the live room
    And user joins a new live room
    And returns "OK" status code
    And "student" receives channel and room ID and other tokens
    When user "raise" hand in the live room
    Then returns "OK" status code
    And user get hands up state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user "folds another learner" hand in the live room
    Then returns "OK" status code
    And user get all learner's hands up states to off in the live room

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    When user "raise" hand in the live room
    Then returns "OK" status code
    And user get hands up state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user "folds all" hand in the live room
    Then returns "OK" status code
    And user get all learner's hands up states to off in the live room
    And have an uncompleted live room log with "3" joined attendees, "4" times getting room state, "4" times updating room state and "0" times reconnection

  Scenario: student try to raise hand and after hand off
    Given student who is part of the live room
    And user joins a new live room
    And returns "OK" status code
    And "student" receives channel and room ID and other tokens
    When user "raise" hand in the live room
    Then returns "OK" status code
    And user get hands up state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user get hands up state in the live room
    Then returns "OK" status code

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    When user "lowers" hand in the live room
    Then returns "OK" status code
    And user get hands up state in the live room
    And have an uncompleted live room log with "2" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection
