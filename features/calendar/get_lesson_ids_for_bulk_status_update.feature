Feature: Get Lesson IDs for Bulk Status Update

    Background:
        Given user signed in as school admin 
        When enter a school
        And have some locations
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have some student subscriptions

    Scenario: User can get lesson IDs for bulk cancel and bulk publish
        Given signed as "school admin" account
        And some existing "<status>" lessons
        When user get lesson IDs for bulk "<action>"
        Then returns "OK" status code
        And returned lesson IDs for bulk "<action>" are expected

        Examples:
            | status     | action   |
            | published  | cancel   |
            | completed  | cancel   |
            | draft      | publish  |