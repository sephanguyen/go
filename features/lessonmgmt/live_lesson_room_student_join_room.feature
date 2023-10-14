Feature: Student join live lesson room

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: students join live lesson
    Given user signed as student who belong to lesson
    When user join live lesson
    Then returns "OK" status code
    And returns valid information for student's broadcast
    And have a uncompleted log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

    When user join live lesson
    Then returns "OK" status code
    And returns valid information for student's broadcast
    And have a uncompleted log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    And returns valid information for student's broadcast
    And have a uncompleted log with "2" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection
