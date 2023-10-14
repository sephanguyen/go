Feature: Modify live lesson room sharing material state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: teacher try to share a material and after stop sharing
    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user share a material with type is video in live lesson room
    Then returns "OK" status code
    And user get current material state of live lesson room is video

    When user share a material with type is pdf in live lesson room
    Then returns "OK" status code
    And user get current material state of live lesson room is pdf

    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    When user get current material state of live lesson room is pdf
    Then returns "OK" status code

    Given user signed in as teacher
    When user stop sharing material in live lesson room
    Then returns "OK" status code
    And user get current material state of live lesson room is empty
    And have a uncompleted log with "2" joined attendees, "4" times getting room state, "3" times updating room state and "0" times reconnection