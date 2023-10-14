Feature: send notification with excluded generic receiver ids
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student
    Scenario Outline: staff with granted role send a notification with some excluded generic receiver ids
        Given current staff upsert notification with valid filter
        When current staff send notification
        Then returns "OK" status code
        And notificationmgmt services must send notification to user
        And excluded user must not receive notification