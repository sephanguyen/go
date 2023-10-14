Feature: Detect text from image

    Scenario Outline: authentication when detect text from image
        Given a signed in "<role>"
        And a list of document images
        When detect text from document images
        Then returns "<status code>" status code
        Examples:
            | role           | status code      |
            | admin          | OK               |
            | school admin   | OK               |
            | hq staff       | OK               |
            | student        | OK               |
            | parent         | PermissionDenied |
            | teacher        | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario Outline: Detect text from image
        Given a signed in "school admin"
        And a list of document images
        When detect text from document images
        Then our system must return texts correctly