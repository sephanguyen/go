Feature: Teacher retrieve stream token
    In order for teacher start live stream
    As a teacher
    I need to retrieve a stream token

    Scenario: teacher retrieve lesson token
        Given a teacher with valid lesson
        When teacher join lesson 
        Then returns "PermissionDenied" status code

    Scenario: teacher from same school retrieve lesson token
        Given a teacher from same school with valid lesson
        And "teacher" signin system
        When teacher join lesson 
        Then returns "OK" status code
        And returns valid information for broadcast

    Scenario: teacher from same school retrieve lesson V1 token
        Given a teacher from same school with valid lesson
        And "teacher" signin system
        When teacher join lesson with v1 API
        Then returns "OK" status code
        And returns valid information for broadcast with v1 API

    Scenario: invalid teacher retrieve lesson token
        Given a teacher with invalid lesson
        When teacher join lesson 
        Then returns "PermissionDenied" status code