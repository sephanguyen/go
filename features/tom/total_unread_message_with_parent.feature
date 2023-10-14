Feature: teacher read all message of parent conversation
    Background: Default resource path
        Given resource path of school "Manabie" is applied

    Scenario: Get total unread message without system messages
        Given a student and conversation created
        And teacher join conversation of "student"
        And teacher read all messages
        Then a parent and conversation created
        When "teacher" get total unread message
        Then returns "OK" status code
        And Tom must returns 0 total unread message