Feature: Export Course Location Schedule
    Background:
        Given user signed in as school admin
        And have some centers
    Scenario Outline: Export all course location schedule
        Given user signed in as school admin
        When user export course location schedule
        Then returns course location schedule in csv with Ok status code
