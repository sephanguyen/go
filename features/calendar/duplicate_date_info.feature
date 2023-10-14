Feature: Duplicate Date Info
    Background:
        Given user signed in as school admin 
        When enter a school
        And an existing location "location_1" in DB
        And a date "2022-06-10", location "location_1", date type "regular", open time "9:30", status "draft"
        And a date "2022-06-15", location "location_1", date type "seasonal", open time "9:30", status "draft"
        And a date "2022-06-22", location "location_1", date type "spare", open time "9:30", status "draft"
        And a date "2022-06-23", location "location_1", date type "closed", open time "nil", status "published"

    Scenario: admin duplicate date info
        Given signed as "school admin" account
        When admin choose "<date>", "<location>"
        And duplicate date info with condition "<condition>", "<from_date>", "<to_date>"
        Then returns "OK" status code

        Examples:
            | date          | location     | condition  | from_date     | to_date       |
            | 2022-06-10    | location_1   | daily      | 2022-06-20    | 2022-06-30    |
            | 2022-06-15    | location_1   | daily      | 2022-06-20    | 2022-06-30    |
            | 2022-06-23    | location_1   | weekly     | 2022-06-20    | 2022-07-20    |

    