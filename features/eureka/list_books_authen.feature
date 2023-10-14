Feature: Authen List Books

    Scenario Outline: Valid and invalid role to list books
        Given some books are existed in DB
        And a signed in "<role>"
        When user list books by ids
        Then returns "<status code>" status code
        Examples:
            | role         | status code |
            | parent       | OK          |
            | student      | OK          |
            | teacher      | OK          |
            | hq staff     | OK          |
            | school admin | OK          |
# | center lead     | OK               |
# | center staff    | PermissionDenied |
# | center manager  | PermissionDenied |
