Feature: Authentication when user assign topic items

    Background: valid book content
        Given a signed in "school admin"
        And a list of valid topics
        And admin inserts a list of valid topics

    Scenario: User assign topic items
        Given a list of valid learning objectives
        And a signed in "<role>"
        When user try to assign topic items with role
        Then returns "<status>" status code
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