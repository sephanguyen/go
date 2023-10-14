Feature: Import time slots

    Background:
        Given some centers

    Scenario: admin try to import time slots csv file success
        Given "staff granted role school admin" signin system
        And a valid time slots payload
        When user import time slots by csv file
        Then returns "OK" status code

    Scenario: admin try to import time slots csv file fail
        Given "staff granted role school admin" signin system
        And an invalid time slots payload
        When user import time slots by csv file
        Then returns "InvalidArgument" status code
