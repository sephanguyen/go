Feature: Insert task assignment

    Background:
        Given <task_assignment>a signed in "school admin"
        And <task_assignment>a valid book content

    Scenario Outline: authenticate <role> when insert task assignment
        Given <task_assignment>a signed in "<role>"
        When user insert a valid task assignment
        Then <task_assignment>returns "<status code>" status code

        Examples:
            | role         | status code      |
            | school admin | OK               |
            | student      | PermissionDenied |
            | parent       | PermissionDenied |
            | teacher      | PermissionDenied |
            | hq staff     | OK               |
    # | center lead    | PermissionDenied |
    # | center manager | PermissionDenied |
    # | center staff   | PermissionDenied |

    Scenario: admin create a task assignment in an existed topic
        Given there are task assignments existed in topic
        When user insert a valid task assignment
        Then task assignment must be created
        And our system generates a correct display order for task assignment
        And our system updates topic LODisplayOrderCounter correctly with new task assignment
