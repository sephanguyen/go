@quarantined
#add quarantined to omit this sample 
Feature: Insert a example

    Background:a valid book content
        Given <example> a signed in "school admin"
        And <example> a valid book content

    Scenario Outline: authenticate insert a example
        Given <example> a signed in "<role>"
        When user insert a example
        Then returns "<msg>" status code

        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | school admin   | OK               |
            | hq staff       | PermissionDenied |
            | teacher        | PermissionDenied |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |

    Scenario: insert a example
        When user insert a example
        Then <example> returns "OK" status code
        And example must be created