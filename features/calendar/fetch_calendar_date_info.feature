Feature: Retrieve Date Info

    Background:
        Given user signed in as school admin 
        When enter a school
        And an existing location "location_1" in DB
        And a date "2022-06-10", location "location_1", date type "regular", open time "9:30", status "draft"
        And a date "2022-06-15", location "location_1", date type "seasonal", open time "9:30", status "draft"
        And a date "2022-06-20", location "location_1", date type "spare", open time "9:30", status "draft"
        And a date "2022-06-21", location "location_1", date type "spare", open time "9:30", status "draft"
        And a date "2022-06-22", location "location_1", date type "spare", open time "9:30", status "draft"
        And a date "2022-06-23", location "location_1", date type "closed", open time "nil", status "published"
        And a date "2022-06-24", location "location_1", date type "regular", open time "9:30", status "draft"
        And a date "2022-06-25", location "location_1", date type "seasonal", open time "9:30", status "draft"
        And a date "2022-06-26", location "location_1", date type "closed", open time "nil", status "draft"

    Scenario: user retrieve calendar with filter by location
        Given signed as "school admin" account
        When user get calendar with filter "<start_date>", "<end_date>", "<location_id>"
        Then returns "OK" status code
        And must return all date info by location

        Examples:
            | start_date    | end_date      | location_id   |
            | 2022-06-20    | 2022-06-30    | location_1    |
            | 2022-06-20    | 2022-06-26    | location_1    |
            | 2022-06-23    | 2022-06-30    | location_1    |