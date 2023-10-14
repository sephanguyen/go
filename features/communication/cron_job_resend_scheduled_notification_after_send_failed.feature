Feature: user send notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App

    @blocker
    Scenario Outline: admin send huge of number sheduled notification to user
        When admin create "<number>" scheduled notification to student
        And waiting for all notification are sent
        Then admin check that "<number>" notification are sent within a minute
        Examples:
            | number |
            | 10     |
            | 50     |

    @blocker
    Scenario: cron job resent failed scheduled notification in next trigger
        Given admin create 2 group of scheduled notification with different time
        When group 1 was sent failed
        Then waiting to group 2 are sent
        And group 1 are also sent
