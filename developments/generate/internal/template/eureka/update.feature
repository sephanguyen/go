Feature: update %[1]s

    Background:
        Given "school admin" logins "CMS"

    Scenario Outline: authenticate when update %[1]s
        Given a signed in "<role>"
        When user update %[1]s
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

    Scenario: update %[1]s
        When user update %[1]s
        Then returns "OK" status code
        And updated %[1]s set as expected

    Scenario: update invalid %[1]s
        Given %[1]s is deleted
        When user update %[1]s
        Then return "NotFound" status code
