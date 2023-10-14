@blocker
Feature: Verify version

    Scenario: Invalid version control
        Given a invalid version request
        When user verify version
        Then returns "InvalidArgument" status code


    Scenario Outline: Need to force update version
        Given a request with lower version "<lower version>"
        When user verify version
        Then returns "OK" status code
        And return false in message

        Examples:
            | lower version      |
            | 0.2.20220831       |
            | 0.5.1              |
            | 1.5.20220923020329 |

    Scenario: Valid version control
        Given a request with valid version
        When user verify version
        Then returns "OK" status code

