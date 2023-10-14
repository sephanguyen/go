Feature: Delete los

    Background:
        Given a signed in "school admin"
        And a list of los created

    Scenario Outline: Valid and invalid role to delete los
        Given a signed in "<role>"
        When user delete los
        Then returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | hq staff       | OK               |
            | teacher        | OK               |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |

    Scenario: delete los
        When user delete los
        Then returns "OK" status code
        And los have been deleted correctly


    Scenario: delete los again
        Given user delete los
        When user delete los again
        Then returns "NotFound" status code