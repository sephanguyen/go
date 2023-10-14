Feature: retrieve classes 
    Background:
        Given some centers
        And have some courses

    Scenario: retrieve classes by ids
        Given "staff granted role school admin" signin system
        And a list of class are existed in DB
        When user retrieve classes
        Then returns "OK" status code
        And must return a correct list of classes
