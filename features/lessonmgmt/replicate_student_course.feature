Feature: Sync student course data from fatima to bob database
    
    Background:
        Given user signed in as school admin
        And have some centers
        And have some teacher accounts
        And have some grades
        And have some student accounts
        And have some courses

    Scenario: sync student course data
        Given user signed in as school admin
        And prepares data for create one time package
        When user creates a student course in order management
        Then returns "OK" status code
        And student course data in bob database has synced successfully
