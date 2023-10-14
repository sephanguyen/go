Feature: Mark info notification as read/unread

    @blocker
    Scenario Outline: user set the notification status
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App
        And school admin upsert notification to student
        And school admin sends notifications to student
        When user set "<status>" the notification
        Then mark the user notification as status "<status>"
        Examples:
            | status                        |
            | USER_NOTIFICATION_STATUS_READ |
            | USER_NOTIFICATION_STATUS_NEW  |
