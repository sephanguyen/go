Feature: Create date info setting
    Background:
        Given user signed in as school admin 
        When enter a school
        And an existing location "12345" in DB
        And an existing date type "regular" in DB

    Scenario Outline: user creates date info setting
        Given signed as "school admin" account
        When user creates a date info for date "2022-06-29" and location "12345"
        Then returns "OK" status code
        And date info is created successfully