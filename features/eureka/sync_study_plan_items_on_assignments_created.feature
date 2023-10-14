@quarantined
Feature: Sync study plan item on assignments created

    Background: Sync study plan item on assignments created background
        Given "school admin" logins "CMS"
        And "teacher" logins "Teacher App"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And "school admin" has created some studyplans exact match with the book content for student

    Scenario: create some assignments
        When user creates some assignments in book
        Then study plan items have created on assignments created correctly 

    Scenario: create some assignments in some books
        When user creates some assignments in books
        Then study plan items have created on assignments created correctly 