Feature: Send scheduled notifications

    cron job will call this api SendScheduledNotifications

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    @blocker
    Scenario:
        Given current staff schedules a notification to be sent within 1 hour
        And current staff schedules a notification to be sent after 1 hour
        When call send scheduled notification within 1 hour
        Then notification scheduled to be sent after 1 hour are not sent
        And notification scheduled to be sent within 1 hour are sent
