Feature: Delete lo study plan item (call internal)

    Background: prepare content book and study plan
        Given "school admin" logins "CMS"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And "school admin" add student to the course
        And "school admin" create study plan from the book

    Scenario Outline: Valid and invalid request to delete lo study plan item
        When "<request>" delete lo study plan item
        Then our system returns "<status code>" status code

        Examples:
            | request |     status code      |
            | valid   |         OK           |
            | invalid |    Unauthenticated   |
    
    Scenario: admin request to delete lo study plan items
        When "valid" delete lo study plan item
        Then our system has to delete lo study plan items correctly