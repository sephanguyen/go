Feature: Create tag
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    @blocker
    Scenario: user create new tag and update successful
        Given tag ID is not existed in database
        When admin upsert tag
        Then tag data is stored in the database correctly
        When admin upsert tag with updated data
        Then tag data is stored in the database correctly
    
    Scenario: user upsert tag but name is exist
        Given tag Name is existed in database
        When admin upsert tag
        Then return error tag name existed
