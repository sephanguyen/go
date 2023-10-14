Feature: Detect formula from image

    Background:
        Given a signed in "school admin"

    Scenario: Detect formula from image
        Given a list of formula images
        When detect formula from images
        Then our system must return formulas correctly