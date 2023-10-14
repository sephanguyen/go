Feature: Delete %[1]s

    Background:
        Given a signed in "school admin"
        And a list of %[1]s created

    Scenario Outline: authenticate when delete %[1]s
        Given a signed in "<role>"
        When user delete %[1]s
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

    Scenario: delete %[1]s
        When user delete %[1]s
        Then returns "OK" status code
        And %[1]s have been deleted correctly

    Scenario: delete %[1]s again
        Given user delete %[1]s
        When user delete %[1]s again
        Then returns "NotFound" status code