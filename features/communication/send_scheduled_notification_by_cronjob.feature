Feature: staff with granted role send notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    @blocker
    Scenario Outline: staff with granted role send notification using all filter
        Given current staff schedules a notification to be sent after 1 minutes
        When waiting for scheduled notification to be sent
        Then notificationmgmt services must send notification to user
        And sent time is valid with scheduled time
