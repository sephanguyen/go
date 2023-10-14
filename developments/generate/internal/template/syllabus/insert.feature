Feature: Insert %[1]s

    Background:
        Given <%[2]s>a signed in "school admin"
        And a valid data in database

    Scenario Outline: authenticate when insert %[1]s
        Given <%[2]s>a signed in "<role>"
        When user insert %[1]s
        Then <%[2]s>returns "<status code>" status code

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

    Scenario: insert %[1]s
        When user insert %[1]s
        Then <%[2]s>returns "OK" status code
        And %[1]s must be created

    Scenario: insert %[1]s with missing fields
        When user insert %[1]s with missing fields
        Then <%[2]s>return "InvalidArgument" status code