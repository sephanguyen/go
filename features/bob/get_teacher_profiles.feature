Feature: Retrieve teacher profiles

    Scenario: student user retrieves teacher profile
        Given "staff granted role teacher" signin system
        And a valid teacher profile
        When user retrieves teacher profile
        Then Bob must returns teacher profile

    Scenario: not found retrieves teacher profile
        Given "staff granted role teacher" signin system
        And a invalid teacher profile
        When user retrieves teacher profile
        Then Bob must returns teacher profile not found