 Feature: Sync date info data from calendar to bob database
    
    Scenario: sync date info data
        Given user signed in as school admin
        When user creates a list of date info to the calendar database
        Then date info data in bob db has synced successfully