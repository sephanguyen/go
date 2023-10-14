Feature: Authentication when user upsert valid quiz

    Background:
        Given a signed in "school admin"

    Scenario Outline: User upsert valid quiz
        Given a signed in "<role>"
        When user upsert a "valid" quiz
        Then returns "<status>" status code
        Examples:
            | role           | status           |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | OK               |
            | student        | PermissionDenied |
            | center lead    | PermissionDenied |
            | parent         | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

