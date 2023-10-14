Feature: Export Zoom Account
    Background:
        Given user signed in as school admin
        And have some centers
        And have some zoom account
    Scenario Outline: Export all zoom account
        Given user signed in as school admin
        When user export zoom accounts
        Then returns zoom accounts in csv with Ok status code
