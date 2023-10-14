Feature: Modify live room raise hands state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: student try to raise hand and after teacher will fold learner's hand
    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user raise hand in live lesson room
    Then returns "OK" status code
    And user get hands up state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user fold a learner's hand in live lesson room
    Then returns "OK" status code
    And user get all learner's hands up states who all have value is off

    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user raise hand in live lesson room
    Then returns "OK" status code
    And user get hands up state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user fold hand all learner
    Then returns "OK" status code
    And user get all learner's hands up states who all have value is off
    And have a uncompleted log with "3" joined attendees, "4" times getting room state, "4" times updating room state and "0" times reconnection

  Scenario: student try to raise hand and after hand off
    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user raise hand in live lesson room
    Then returns "OK" status code
    And user get hands up state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user get hands up state
    Then returns "OK" status code

    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user hand off in live lesson room
    Then returns "OK" status code
    And user get hands up state
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection
