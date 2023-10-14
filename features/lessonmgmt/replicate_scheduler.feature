 Feature: Sync scheduler data from calendar to bob database
    
    Scenario: sync scheduler data
        Given user signed in as school admin
        When user creates a scheduler to the calendar database
        Then scheduler data in bob db has synced successfully