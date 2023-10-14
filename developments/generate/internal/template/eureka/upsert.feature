Feature: Upsert %[1]s

    Background:
        Given "school admin" logins "CMS"

    Scenario Outline: authenticate when upsert %[1]s
        Given a signed in "<role>"
        When user upsert %[1]s
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

    Scenario: create %[1]s
        When user create %[1]s
        Then returns "OK" status code
        And %[1]s must be created

    Scenario: update %[1]s
        Given user create %[1]s
        When user update %[1]s
        Then returns "OK" status code
        And %[1]s must be updated
