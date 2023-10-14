Feature: Export Classes

    Export classes masterdata

    Scenario Outline: Export classes
        Given "school admin" signin system
        And classes existed in DB
        When user export classes
        Then returns classes in csv with Ok status code

