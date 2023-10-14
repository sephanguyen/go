@cms @learner @parent
@communication
@scheduled-notification
@ignore

@wip
Feature: Update status of scheduled notification on CMS

    Background:
        Given "school admin" logins CMS
        And "school admin" has created a student with grade, course and parent info
        And "student" logins Learner App
        And "school admin" is at "Notification" page on CMS
        And "school admin" has created a scheduled notification

    Scenario: Update a scheduled notification to draft notification
        Given "school admin" has opened editor full-screen dialog of scheduled notification
        When "school admin" selects status "Now"
        And "school admin" clicks "Save draft" button
        Then "school admin" sees scheduled notification has been saved to draft notification

    Scenario: Update a scheduled notification to sent notification
        When "school admin" waits for scheduled notification to be sent on time
        Then status of scheduled notification is updated to "Sent