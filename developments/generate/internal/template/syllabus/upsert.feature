Feature: Upsert %[1]s

    Background:
        Given <%[2]s>a signed in "school admin"

    Scenario Outline: authenticate when upsert %[1]s
        Given <%[2]s>a signed in "<role>"
        When user upsert %[1]s
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

    Scenario: create %[1]s
        When user create %[1]s
        Then <%[2]s>returns "OK" status code
        And %[1]s must be created

    Scenario: update %[1]s
        Given user create %[1]s
        When user update %[1]s
        Then <%[2]s>returns "OK" status code
        And %[1]s must be updated
