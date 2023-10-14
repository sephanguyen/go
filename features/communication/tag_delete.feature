@quarantined
Feature: school admin deletes a tag
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And a valid delete tag request
    Scenario: school admin delete a tag
        Given tag is exist in database
        When admin delete tag
        Then tag is soft delete in database