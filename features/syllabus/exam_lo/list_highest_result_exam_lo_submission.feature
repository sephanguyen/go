Feature: List highest result exam LO submission

    Scenario Outline: authenticate when list exam LO
        Given <exam_lo>a signed in "<role>"
        And there are exam lo submissions existed
        When user list highest result exam LO submission
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role         | status code |
            | school admin | OK          |
            | admin        | OK          |
            | teacher      | OK          |
            | student      | OK          |
            | hq staff     | OK          |

    Scenario Outline: authenticate when list exam LO
        Given <exam_lo>a signed in "teacher"
        And there are exam lo submissions existed
        When user list highest result exam LO submission
        Then <exam_lo>returns "OK" status code
        And our system must return highest result exam lo submissions correctly