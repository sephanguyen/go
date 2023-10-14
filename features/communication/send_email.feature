Feature: send internal email
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        Then waiting for kafka sync user info for notificationmgmt database

    Scenario: staff send an email successfully
        When current staff send an email
        Then returns "OK" status code
        And spike service must save this email and email recipients
