Feature: Leave live lesson
    In order for user to see live stream
    As a user
    I need to leave lesson

    Background:
        Given some teacher accounts with school id
        And some student accounts with school id
        And some live courses with school id
        And some medias
        And a live lesson

    Scenario: student leave lesson
        Given a signed in teacher
        When user share a material with type is pdf in live lesson room
        Then returns "OK" status code
        And user get current material state of live lesson room is pdf

        Given user signed as student who belong to lesson
        When student leave lesson
        Then returns "OK" status code

    Scenario: student leave lesson for other student
        Given a signed in student
        When student leave lesson for other student
        Then returns "PermissionDenied" status code
