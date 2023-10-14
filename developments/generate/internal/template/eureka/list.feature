Feature: List %[1]s

    Background:
        Given a signed in "school admin"
        And a list of %[1]s created

    Scenario Outline: authenticate when list %[1]s
        Given a signed in "<role>"
        When user list %[1]s
        Then returns "<status code>" status code

        Examples:
            | role           | status code |
            | school admin   |             |
            | admin          |             |
            | teacher        |             |
            | student        |             |
            | hq staff       |             |
            | center lead    |             |
            | center manager |             |
            | center staff   |             |
            | lead teacher   |             |

    Scenario: list %[1]s
        When user list %[1]s
        Then returns "OK" status code
        And our system must return %[1]s correctly

    Scenario: list invalid %[1]s
        Given %[1]s is deleted
        When user list %[1]s 
        Then returns "NotFound" status code