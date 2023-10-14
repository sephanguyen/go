Feature: Authen Upsert Books

    Background:
        Given a signed in "school admin"

    Scenario Outline: Valid and invalid role to upsert books
        Given a signed in "<role>"
        When user upsert valid books
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
