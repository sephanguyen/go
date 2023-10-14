Feature: Modify live lesson room polling state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: teacher try start polling and after stop, end it
    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user start polling in live lesson room
    Then returns "OK" status code
    And user get current polling state of live lesson room started

    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user submit the answer "A" for polling
    Then returns "OK" status code
    And user get polling answer state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user stop polling in live lesson room
    Then returns "OK" status code
    And user get current polling state of live lesson room stopped

    When user end polling in live lesson room
    Then returns "OK" status code
    And user get current polling state of live lesson room is empty

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user start polling in live lesson room
    Then returns "OK" status code
    And user get current polling state of live lesson room started

    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user submit the answer "B" for polling
    Then returns "OK" status code
    And user get polling answer state
    And have a uncompleted log with "4" joined attendees, "6" times getting room state, "6" times updating room state and "0" times reconnection
