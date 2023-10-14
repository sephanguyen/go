Feature: List assignment

    Background:
        Given <assignment>a signed in "school admin"
        And <assignment>a valid book content
        And there are assignments existed

    Scenario Outline: authenticate when list assignment
        Given <assignment>a signed in "<role>"
        When user list assignment
        Then <assignment>returns "<status code>" status code

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

    Scenario: list vaild assignment
        When user list assignment
        Then <assignment>returns "OK" status code
        And our system must return assignments correctly