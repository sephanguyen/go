@quarantined
Feature: Teacher retrieve stream token
    In order for teacher start live stream
    As a teacher
    I need to retrieve a stream token

    Scenario: teacher from same school retrieve lesson token
        Given a teacher from same school with valid lesson
        When teacher retrieve stream token
        Then returns "OK" status code

    Scenario: invalid teacher retrieve lesson token
        Given a teacher with invalid lesson
        When teacher retrieve stream token
        Then returns "PermissionDenied" status code
