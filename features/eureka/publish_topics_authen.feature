Feature: Authentication when user public topics

    Background:
        Given a signed in "school admin"
        And user has created an empty book
        And user create a valid chapter
        And user has created some valid topics

    Scenario Outline: Uer public topics
        Given a signed in "<role>"
        When user public some topics
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