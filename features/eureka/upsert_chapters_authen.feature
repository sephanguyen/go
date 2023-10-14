Feature: Authen Upsert Chapters

    Background:
        Given a signed in "school admin"
        And user has created an empty book

    Scenario Outline: Valid and invalid role to upsert chapters
        Given a signed in "<role>"
        When user upsert valid chapters
        Then returns "<status code>" status code
        Examples:
            | role           | status code      |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | teacher        | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |
            | hq staff       | OK               |
            | school admin   | OK               |
