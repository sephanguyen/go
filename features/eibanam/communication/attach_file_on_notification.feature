@cms @learner @parent
@communication

Feature: Attach file

    Background:
        Given "school admin" logins CMS
        And "school admin" has created 1 course
        And "school admin" has created a student with grade and parent info
        And "school admin" has added created course for student
        And "student" logins Learner app
        And "parent" logins Learner app
        And "school admin" is at "Notification" page on CMS

    Scenario: Create and send notification with a PDF file
        Given school admin has saved a draft notification with a PDF file size smaller than 50 MB
        When school admin sends notification
        Then school admin sends notification successfully
        And "student" receives the notification in their device
        And "parent" receives the notification in their device
