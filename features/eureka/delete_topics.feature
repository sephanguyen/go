Feature: Delete topics
    Background:
        Given a signed in "school admin"
        And user has created an empty book
        And user create a valid chapter
        And user has created some valid topics

    Scenario: User delete valid topics
        When user delete some topics
        Then our system must delete the topics correctly

    Scenario: User delete some non-existent topics
        Given some missing topic ids
        When user delete some topics
        Then returns "InvalidArgument" status code

    Scenario: User delete some deleted topics
        Given user delete some topics
        When user delete some topics
        Then returns "InvalidArgument" status code
