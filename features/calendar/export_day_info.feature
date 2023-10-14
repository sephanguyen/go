Feature: Export Day Infos
    Background:
        Given user signed in as school admin 
        When enter a school
        And an existing location "location_1" in DB
        And a date "2022-11-10", location "location_1", date type "regular", open time "9:30", status "draft"
        And a date "2022-11-15", location "location_1", date type "seasonal", open time "9:30", status "draft"
        And a date "2022-11-22", location "location_1", date type "spare", open time "9:30", status "draft"
        And a date "2022-11-23", location "location_1", date type "closed", open time "nil", status "published"
    
    Scenario Outline: Export day infos
        Given signed as "school admin" account
        When user export day infos
        Then returns day infos in csv with Ok status code
