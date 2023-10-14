Feature: Execute script

    Scenario: admin execute script custom entity
        Given "staff granted role school admin" signin system
        When school admin execute script
        Then returns "OK" status code