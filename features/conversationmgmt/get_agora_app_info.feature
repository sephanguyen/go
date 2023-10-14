Feature: test get agora app info
    Scenario: get agora app info successful
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And current staff creates "1" students
        When student login to Learner App
        Then student call GetAgoraInfo API
        Then returns "OK" status code
        And GetAgoraInfo API return correct data
