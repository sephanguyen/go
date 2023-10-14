@cms @learner
@communication

Feature: Create and send notification

    Background:
        Given "school admin" logins CMS
        And "school admin" has created 1 course
        And "school admin" has created a student with grade and parent info
        And "school admin" has added created course for student
        And "student" logins Learner app
        And "parent" logins Learner app
        And "school admin" is at "Notification" page on CMS

    Scenario: Create and send notification with required fields
        When "school admin" sends notification with required fields to student and parent
        Then school admin sends notification successfully
        And "student" receives the notification in their device
        And "parent" receives the notification in their device

    Scenario: Create and send a draft notification
        Given school admin has saved a draft notification with required fields
        # instruction
        # And school admin has selected created course in course, grade of created student in grade for draft notification
        # And school admin has selected created student email in individual recipient for draft notification
        # And school admin has selected "All" in type
        When school admin sends that draft notification for student and parent
        Then school admin sends notification successfully
        And "student" receives the notification in their device
        And "parent" receives the notification in their device