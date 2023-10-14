@runsequence

Feature: Export location

    Export location masterdata

    Scenario Outline: Export location
        Given "staff granted role school admin" signin system
        And some locations existed in DB
        When user export locations
        Then returns locations in csv with Ok status code
