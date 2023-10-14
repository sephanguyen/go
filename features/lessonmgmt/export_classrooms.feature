Feature: Export Classrooms
    Background:
        Given user signed in as school admin
        And have some centers
        And have some classrooms

    Scenario Outline: Export all classrooms
        Given user signed in as school admin
        When user export classrooms
        Then returns "OK" status code
        And returns classrooms in csv with "location_id,location_name,classroom_id,classroom_name,remarks,is_archived" columns
