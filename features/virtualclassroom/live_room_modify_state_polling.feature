Feature: Modify Live Room Polling

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

  Scenario: teacher try to start, stop, share and end polling with learner to submit polling answer
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And "teacher" receives channel and room ID and other tokens
    When user start polling with "<number_options>" options and "<number_correct>" correct answers in the live room
    Then returns "OK" status code
    And user get current polling state in the live room has started

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    When user submit the answer "A,B" in the live room polling
    Then returns "OK" status code
    And user get polling answer state in the live room

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user stop polling in the live room
    Then returns "OK" status code
    And user get current polling state in the live room has stopped

    When user "started" sharing the polling in the live room
    Then returns "OK" status code
    And user get current polling state of the live room containing "started" share polling

    When user "stopped" sharing the polling in the live room
    Then returns "OK" status code
    And user get current polling state of the live room containing "stopped" share polling

    When user end polling in the live room
    Then returns "OK" status code
    And user get current polling state in the live room is empty

    Given "teacher" signin system
    And user joins an existing live room
    And returns "OK" status code
    When user start polling in the live room
    Then returns "OK" status code
    And user get current polling state in the live room has started

    Given student who is part of the live room
    And user joins an existing live room
    And returns "OK" status code
    When user submit the answer "B,C" in the live room polling 
    Then returns "OK" status code
    And user get polling answer state in the live room
    And have an uncompleted live room log with "4" joined attendees, "8" times getting room state, "8" times updating room state and "0" times reconnection

    Examples:
      | number_options | number_correct |
      | 2              | 0              |
      | 5              | 0              |
      | 10             | 0              |
      | 2              | 2              |
      | 10             | 10             |