Feature: Monitor upsert course student (Check study plan for student according an arbitrary course)

    Background: prepare content book and study plan
        Given "school admin" logins "CMS"
        And "school admin" has created a content book
        And "school admin" add some students to the course
        And "school admin" create study plan from the book

    Scenario: Some missing student study plans
        Given some study plan items not created 
        When run monitor upsert learning item
        Then our monitor save missing learning item correctly
        Then our monitor auto upsert missing learning item correctly    