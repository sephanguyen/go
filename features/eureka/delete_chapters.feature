Feature: Delete chapters
    Background:
        Given a signed in "school admin"
        And user has created an empty book
        And there are chapters existed

    Scenario: User delete valid chapters
        When user delete some chapters
        Then our system must delete the chapters correctly

    Scenario: User delete some non-existent chapters
        Given some missing chapter ids
        When user delete some chapters
        Then returns "InvalidArgument" status code

    Scenario: User delete some deleted chapters
        Given user delete some chapters
        When user delete some chapters
        Then returns "InvalidArgument" status code
