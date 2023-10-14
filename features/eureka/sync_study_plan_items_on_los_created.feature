Feature: Sync study plan item on los created

    Background: Sync study plan item on los created background
        Given "school admin" logins "CMS"
        And "teacher" logins "Teacher App"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And "school admin" has created some studyplans exact match with the book content for student

    @quarantined
    Scenario: create some los
        When user creates some los in book
        Then study plan items have created correctly

    @quarantined
    Scenario: create some los in some books
        Given "school admin" has created some studyplans exact match with some books content for student
        When user creates some los in books
        Then study plan items have created correctly
