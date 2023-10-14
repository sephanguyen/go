Feature: write comment for student
    In write and view comment for student
    I need to upsert and view comment for student

    Scenario: a student write comment for a student
        Given a signed in student
        And a user comment for his student
        When user upsert comment for his student
        Then returns "PermissionDenied" status code

    Scenario: a admin write comment for a student
        Given a signed in student
        And a user comment for his student
        When user upsert comment for his student
        Then returns "PermissionDenied" status code

    Scenario: a student retrieve comment of a student
        Given a signed in student
        And valid comment for student in DB
        When user retrieve comment for student
        Then returns "PermissionDenied" status code

    Scenario: a teacher write comment for a student
        Given a signed in student
        Given a signed in teacher
        And a user comment for his student
        When user upsert comment for his student
        Then Bob must store comment for student
        And user retrieve comment for student
        And Bob must return all comment for student


    Scenario: a teacher retrieves comments for a student
        Given a signed in student
        Given a signed in teacher
        And a teacher gives some comments for student
        When the teacher retrieves comment for student
        Then our system have to response retrieve comment correctly

