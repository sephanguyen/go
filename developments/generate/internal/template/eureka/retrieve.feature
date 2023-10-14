Feature: retrieve %[1]s

    Background:
        Given a signed in "school admin"
        And a valid data in database

    Scenario Outline: authenticate when retrieve %[1]s
        Given a signed in "<role>"
        When user try %[1]s
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

    Scenario: retrieve %[1]s
        When user try %[1]s
        Then returns "OK" status code
        And our system must return results correctly

    Scenario: retrieve invalid %[1]s
        Given %[1]s is deleted
        When user try %[1]s
        Then returns "NotFound" status code