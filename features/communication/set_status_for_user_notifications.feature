Feature: Mark info notification as read/unread

    Scenario Outline: user set the notification status
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student "0" logins to Learner App
        And school admin sends some notificationss to a student
        When user set "<status>" the notification
        Then mark the user notification as status "<status>"
        Examples:
            | status                        |
            | USER_NOTIFICATION_STATUS_READ |
            | USER_NOTIFICATION_STATUS_NEW  |
