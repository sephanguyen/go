@quarantined
Feature: Retrieve Whiteboard Token

    Background:
        Given a signed in "school admin"
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
        And admin inserts schools

        Given a signed in teacher

    Scenario: student retrieve whiteboard token
        Given a signed in student
        And a list of courses are existed in DB of "above teacher"
        And a student with valid lesson
        When retrieve whiteboard token
        Then returns "OK" status code
        And receive whiteboard token

    Scenario: student retrieve whiteboard token which has no room id
        Given a signed in student
        And a list of courses are existed in DB of "above teacher"
        And a student with valid lesson which has no room id
        When retrieve whiteboard token
        Then returns "OK" status code
        And receive whiteboard token

        When retrieve whiteboard token
        Then returns "OK" status code
        And receive whiteboard token

    Scenario: student retrieve whiteboard token without permission
        Given a signed in student
        And a list of courses are existed in DB of "above teacher"
        When retrieve whiteboard token
        Then returns "PermissionDenied" status code

    Scenario: teacher retrieve whiteboard token
        Given a teacher with valid lesson
        When retrieve whiteboard token
        Then returns "OK" status code
        And receive whiteboard token
