Feature: Export Course teaching time
    Background:
        Given user signed in as school admin
        And have some student accounts
        And have some courses
        And have some student subscriptions

    Scenario Outline: Export all Courese teaching time
        Given user signed in as school admin
        And register some course's teaching time
        When user export course with teaching time info
        Then returns "OK" status code
        And returns course with teaching time info in csv
