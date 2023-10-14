Feature: update learning objective name
    Background:
        Given a signed in "school admin"
        And a valid book content
        And user create learning objectives

    Scenario: update learning objective name
        When user update learning objective name
        Then our system must update learning objective name correctly
