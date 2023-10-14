Feature: Update date info setting
    Background:
        Given user signed in as school admin 
        When enter a school
        And an existing location "12345" in DB
        And an existing date type "regular" in DB
        And an existing date type "spare" in DB
        And an existing date info for date "2022-06-30" and location "12345" 

    Scenario Outline: user updates date info setting
        Given signed as "school admin" account
        When user updates date info for date "2022-06-30" and location "12345"
        Then returns "OK" status code
        And date info is updated successfully