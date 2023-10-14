Feature: Retrieve Locations for Academic Calendar

    Background:
        Given some centers
        And some location types
        And have academic year

    Scenario: admin try to get locations for academic calendar import
        Given "staff granted role school admin" signin system
        When user retrieve locations for academic calendar
        Then returns "OK" status code
