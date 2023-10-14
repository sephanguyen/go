Feature: Authentication when user upsert topics

    Background:
        Given a signed in "school admin"
        And user has created an empty book
        And user create a valid chapter

    Scenario Outline: User upsert topics
        Given a signed in "<role>"
        When user has created some "type" topics
        Then  returns "<status>" status code
        Examples:
            | role           | status           |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | center lead    | PermissionDenied |
            | parent         | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |
