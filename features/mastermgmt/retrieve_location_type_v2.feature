@quarantined
Feature: retrieve location types (v2)

    Scenario: retrieve location types (v2)
        Given "staff granted role school admin" signin system
        And a valid location type payload
        Then user import location type by csv file
        When user retrieve location types v2
        Then returns "OK" status code
        And returns unArchived location types
