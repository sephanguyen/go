@quarantined
Feature: Monitor upsert course student (Check study plan for student according an arbitrary course)

    Background: prepare content book and study plan
        Given "school admin" logins "CMS"
        And "school admin" has created a content book
        And "school admin" add some students to the course
        And "school admin" create study plan from the book

    Scenario: Some missing student study plans
        Given some student's study plans not created 
        When run monitor upsert course student
        Then our monitor save missing student correctly
    