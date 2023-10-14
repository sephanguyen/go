Feature: Export location types v2

    Export location types masterdata v2
    Scenario Outline: Export location types v2
        Given "school admin" signin system
        And a valid location type payload
        Then user import location type by csv file
        When user export location types
        Then returns "v2" location types in csv with Ok status code
