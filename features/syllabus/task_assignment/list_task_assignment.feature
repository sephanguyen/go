Feature: List task_assignment

    Background:
        Given <task_assignment>a signed in "school admin"
        And <task_assignment>a valid book content
        And there are task assignments existed in topic

    Scenario Outline: authenticate when list task_assignment
        Given <task_assignment>a signed in "school admin"
        When user list task assignment
        Then <task_assignment>returns "<status code>" status code

        Examples:
            | role           | status code |
            | school admin   | OK          |
            | admin          | OK          |
            | teacher        | OK          |
            | student        | OK          |
            | hq staff       | OK          |
            | center lead    | OK          |
            | center manager | OK          |
            | center staff   | OK          |
            | lead teacher   | OK          |

    Scenario: list task_assignment
        When user list task assignment
        Then <task_assignment>returns "OK" status code
        And our system must return task assignment correctly
