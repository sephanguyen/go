Feature: retrieve location types

    Scenario: retrieve location types
        Given "staff granted role school admin" signin system
        And a valid location type payload
        Then user import location type by csv file
        When user retrieve location types
        Then returns "OK" status code
        And must return a correct list of location types
