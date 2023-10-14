Feature: update task assignment

    Background:
        Given <task_assignment>a signed in "school admin"
        And <task_assignment>a valid book content
        And there are task assignments existed in topic

    Scenario Outline: authenticate <role> when update task assignment
        Given <task_assignment>a signed in "<role>"
        When user update valid task assignment
        Then <task_assignment>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | OK               |
            | center staff   | PermissionDenied |
            | lead teacher   | PermissionDenied |

    Scenario: update valid task assignment
        When user update valid task assignment
        Then <task_assignment>returns "OK" status code
        And  our system updates the task assignment correctly
