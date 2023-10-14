Feature: update %[1]s

    Background:
        Given <%[2]s>a signed in "school admin"

    Scenario Outline: authenticate when update %[1]s
        Given <%[2]s>a signed in "<role>"
        When user update %[1]s
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

    Scenario: update %[1]s
        When user update %[1]s
        Then <%[2]s>returns "OK" status code
        And updated %[1]s set as expected

    Scenario: update invalid %[1]s
        Given %[1]s is deleted
        When user update %[1]s
        Then <%[2]s>return "NotFound" status code
