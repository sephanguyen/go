Feature: Update student

    Background: update student
        Given "school admin" logins "CMS"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And user create a course with a study plan

    Scenario: study plan have stored correctly
        When user add course to student
        Then study plan of student have stored correctly