@quarantined
Feature: Import study plan edge cases

    Background: create new course
        Given "school admin" logins
        And "teacher" logins
        And "student" logins

    Scenario: Import two books into two courses
        Given school admin create two books and import two books into study plan
        When user create new los and assignments
        Then new los and assignments must be created
        And study plan items wrong book_id

    Scenario: Import one book into two courses
        Given school admin create one book and import one book into two courses
        When user create new los and assignments
        Then new los and assignments must be created
    
