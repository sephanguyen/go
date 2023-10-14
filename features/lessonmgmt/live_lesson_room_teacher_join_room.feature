Feature: Teacher join live lesson room

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

#  Scenario: teacher retrieve lesson token
#    Given a teacher with valid lesson
#    When teacher join lesson
#    Then returns "PermissionDenied" status code

  Scenario: teachers from same school join live lesson
    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    And returns valid information for broadcast
    And have a uncompleted log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

    When user join live lesson
    Then returns "OK" status code
    And returns valid information for broadcast
    And have a uncompleted log with "1" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    And returns valid information for broadcast
    And have a uncompleted log with "2" joined attendees, "0" times getting room state, "0" times updating room state and "0" times reconnection

#  Scenario: teacher from same school retrieve lesson V1 token
#    Given a teacher from same school with valid lesson
#    When teacher join lesson with v1 API
#    Then returns "OK" status code
#    And returns valid information for broadcast with v1 API

#  Scenario: invalid teacher retrieve lesson token
#    Given a teacher with invalid lesson
#    When teacher join lesson
#    Then returns "PermissionDenied" status code