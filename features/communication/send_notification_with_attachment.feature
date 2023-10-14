Feature: Admin create media attachment notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students with "1" parents info for each student
        And school admin creates "1" courses
        And school admin add packages data of those courses for each student
        And student "0" logins to Learner App
        And update user device token to an "valid" device token

    @blocker
    Scenario: admin create immediate media attachment notification
        Given admin create media attachment notification and sent
        Then student receive media attachment notification
