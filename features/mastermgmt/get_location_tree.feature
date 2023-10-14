@quarantined
Feature: get location tree

    Scenario: get location tree
        Given "school admin" signin system
        And locations with children existed
        When user gets location tree
        Then returns "OK" status code
        And must return a correct location tree
