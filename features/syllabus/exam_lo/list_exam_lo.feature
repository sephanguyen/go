Feature: List exam LO

    Background:
        Given <exam_lo>a signed in "school admin"
        And <exam_lo>a valid book content
        And there are exam LOs existed in topic

    Scenario Outline: authenticate when list exam LO
        Given <exam_lo>a signed in "<role>"
        When user list exam LOs
        Then <exam_lo>returns "<status code>" status code

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

    Scenario: list exam LO
        When user list exam LOs
        Then <exam_lo>returns "OK" status code
        And our system must return exam LOs correctly

    Scenario: list exam LO has a total question
        Given a valid quiz set for exam LO
        When user list exam LOs
        And our system must return exam LOs has a total question
