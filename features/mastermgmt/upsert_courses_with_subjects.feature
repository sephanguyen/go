Feature: Upsert courses
    Background:
        Given a random number
        And some centers
        And some course types

    Scenario Outline: user update/insert courses with subjects
        Given "staff granted role school admin" signin system
        And some subjects
        When user upsert courses with subjects "<subjects>"
        Then returns "OK" status code
        And courses updated with correct subjects
        Examples:
            | subjects |
            | all_new  |
            | add_new  |
            | modified |
            | delete   |