Feature: Import working hours

    Background: 
        Given some centers

    Scenario: admin try to import working hours csv file success
        Given "staff granted role school admin" signin system
        And a valid working hours payload
        When user import working hours by csv file
        Then returns "OK" status code

    Scenario: admin try to import working hours csv file fail
        Given "staff granted role school admin" signin system
        And an invalid working hours payload
        When user import working hours by csv file
        Then returns "InvalidArgument" status code