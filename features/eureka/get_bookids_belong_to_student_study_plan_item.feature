Feature: Get book ids which belong to student study plan items (call internal)

    Background: prepare content book and study plan
        Given "school admin" logins "CMS"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And "school admin" add student to the course
        And "school admin" create study plan from the book

    Scenario Outline: Valid and invalid request to get book ids which belong to student study plan items
        When "<request>" get book ids belong to student study plan items
        Then our system returns "<status code>" status code

        Examples:
            | request |     status code      |
            | valid   |         OK           |
            | invalid |    Unauthenticated   |
    
    Scenario: admin request to get book ids belong to student stydy plan items
        When "valid" get book ids belong to student study plan items
        Then our system has to get book ids belong to student study plan items correctly