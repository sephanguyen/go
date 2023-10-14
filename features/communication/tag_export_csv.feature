Feature: school admin exports tag with csv
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin create some tags named "tag1,tag2,tag3,tag4,tag5"
    
    @blocker
    Scenario: admin export tags
        When school admin export tags
        Then csv data is correctly exported
