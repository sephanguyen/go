Feature: school admin imports tag with csv
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
    Scenario: admin import csv to empty database
        Given a valid csv file
        When admin import csv tag file
        Then csv data is correctly stored in database

    Scenario: admin import csv to exist database
        Given admin create "random" tag with "random" keywords
        And a valid csv file
        When admin import csv tag file
        Then csv data is correctly stored in database

    @blocker
    Scenario: admin import csv to upsert and insert new data
        Given admin create "random" tag with "random" keywords
        And a valid csv file
        When admin import csv tag file
        Then csv data is correctly stored in database
        When admin update csv file
        And admin import csv tag file
        Then csv data is correctly stored in database

    Scenario: admin import invalid csv
        Given a invalid csv file with "<type>"
        When admin import csv tag file
        Then returns "InvalidArgument" status code and error message have "<err_message>"
        Examples:
            | type           | err_message                                                                 |
            | wrong header   | Header "tag_order" is not allowed. Only allow tag_id, tag_name, is_archived |
            | missing header | Missing headers "is_archived"                                               |

    Scenario: admin try to create invalid data
        Given admin create "random" tag with "random" keywords
        And a valid csv file
        When admin update csv to "<err_type>" tag "<tag_field>"
        Then admin import csv tag file
        Then returns "InvalidArgument" status code
        Examples:
            | err_type  | tag_field |
            | duplicate | name      |
            | duplicate | id        |
            | not exist | id        |
