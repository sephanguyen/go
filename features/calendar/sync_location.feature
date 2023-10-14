 Feature: Sync location data from bob to calendar database

    Scenario Outline: sync location table 
        Given user signed in as school admin 
        When a list of location records has been added to the bob database
        Then location data on calendar db updated successfully

    Scenario Outline: sync location type table 
        Given user signed in as school admin 
        When a list of location type records has been added to the bob database
        Then location type data on calendar db updated successfully