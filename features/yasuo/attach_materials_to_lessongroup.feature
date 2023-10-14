Feature: Attach meterial to lesson group
    Background: User login as admin
        Given a signed in "school admin"
        
    Scenario: Materials will be attached to lesson group
        Given a valid course
        When admin attach materials into lesson group
        Then system must attach materials into lesson group correctly