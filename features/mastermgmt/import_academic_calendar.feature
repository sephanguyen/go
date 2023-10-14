Feature: Import academic calendar

    Background:
        Given some centers
        And have academic year

    Scenario: admin try to import academic calendar csv file success
        Given "staff granted role school admin" signin system
        And a valid academic calendar payload
        When user import academic calendar by csv file
        Then returns "OK" status code

    Scenario: admin try to import academic calendar csv file fail
        Given "staff granted role school admin" signin system
        And an invalid academic calendar payload
        When user import academic calendar by csv file
        Then returns "InvalidArgument" status code
