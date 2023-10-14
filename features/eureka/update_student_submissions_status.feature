Feature: Update student submissions status

    Background: Update student submissions status
        Given "school admin" logins "CMS"
        And a list students logins "Learner App"
        And "teacher" logins "Teacher App"

    Scenario: Update student submissions status
        Given "school admin" create a study plan with book have an assignment
        And "school admin" add students to course
        And students do assignment
        And "teacher" grade submissions with status in progress
        When "teacher" update student submissions status to returned
        Then our system returns "OK" status code
        And notifications has been stored correctly