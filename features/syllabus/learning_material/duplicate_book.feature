Feature: Duplicate book

    Background:
        Given <learning_material>a signed in "school admin"
        And a valid book in db

    Scenario Outline: authenticate when duplicate book
        Given <learning_material>a signed in "<role>"
        When user send duplicate book request
        Then <learning_material>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | teacher        | OK               |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | lead teacher   | OK               |

    Scenario: Copy all content in book
        When user send duplicate book request
        Then <learning_material>returns "OK" status code
        And our system must return copied book correctly


