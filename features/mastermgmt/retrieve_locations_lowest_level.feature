@quarantined
Feature: retrieve lowest level locations

    Scenario Outline: retrieve lowest level locations
        Given "staff granted role school admin" signin system
        And a random number in range 999999
        And a valid location type payload
        Then user import location type by csv file
        Then returns "OK" status code
        And a list of locations with variant types are existed in DB
        When user retrieve lowest level of locations with filter "<locationIDs>"
        Then returns "OK" status code
        And must return lowest level of locations with filter "<locationIDs>"

        Examples:
            | locationIDs                 |
            | location-id-8,location-id-9 |

    Scenario Outline: retrieve lowest level locations v2
        Given "staff granted role school admin" signin system
        And a random number in range 999999
        And a valid location type payload
        Then user import location type by csv file
        Then returns "OK" status code
        And a list of locations with variant types are existed in DB
        When user retrieve lowest level of locations with filter "<locationIDs>"
        Then returns "OK" status code
        And must return lowest level of locations with filter "<locationIDs>"

        Examples:
            | locationIDs                 |
            | location-id-8,location-id-9 |
