@quarantined
Feature: Modify live lesson room recording state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: teacher try request recording and after stop
    Given "first" user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user request recording live lesson
    Then returns "OK" status code
    And user get current recording live lesson permission to start recording

    Given "second" user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user request recording live lesson
    Then returns "OK" status code
    And user have no current recording live lesson permission

    Given "second" user signed in as teacher
    When user stop recording live lesson
    Then returns "OK" status code
    And live lesson is still recording

    Given "first" user signed in as teacher
    When user stop recording live lesson
    Then returns "OK" status code
    And live lesson is not recording

    Given "second" user signed in as teacher
    When user request recording live lesson
    Then returns "OK" status code
    And user get current recording live lesson permission to start recording
    And have a uncompleted log with "2" joined attendees, "5" times getting room state, "5" times updating room state and "0" times reconnection