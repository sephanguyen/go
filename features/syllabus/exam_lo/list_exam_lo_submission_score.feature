Feature: List exam lo submission score

    Scenario Outline: authenticate when list exam lo submission score
        Given <exam_lo>a signed in "<role>"
        And there are exam lo submission scores existed
        When user list exam lo submission scores
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role         | status code |
            | school admin | OK          |
            | admin        | OK          |
            | teacher      | OK          |
            | student      | OK          |
            | hq staff     | OK          |

    Scenario: list exam lo submission score
        Given <exam_lo>a signed in "teacher"
        And there are exam lo submission scores existed
        When user list exam lo submission scores
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submission scores correctly
    
    Scenario: list exam lo submission score with question groups
        Given <exam_lo>a signed in "school admin"
        And insert a question group
        And <exam_lo>a signed in "teacher"
        And there are exam lo submission scores existed
        When user list exam lo submission scores
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submission scores correctly