Feature: get list staff under a location

    Background:
        Given user signed in as school admin 
        When enter a school
        And have some locations

    Scenario: user get list staff under location without user group
        Given signed as "school admin" account
        And a list of staff already exists in DB
        When user get list staff
        Then returns "OK" status code
        And an empty list staff is returned

    Scenario: user get list staff under location with user group
        Given signed as "school admin" account
        And a list of staff with user group already exists in DB
        When user get list staff
        Then returns "OK" status code
        And a list correct staff is returned
