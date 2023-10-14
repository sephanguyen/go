Feature: Store device token
    Scenario: student try to store device token
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student "0" logins to Learner App
        When student try to store device token
        Then user's device token is stored to DB
        And NotificationMgmt must publish event to user_device_token channel
