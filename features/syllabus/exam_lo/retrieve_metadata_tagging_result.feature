Feature: Retrieve metadata tagging result

    Background:
        Given <exam_lo>a signed in "school admin"
        And <exam_lo>a valid book content

    Scenario Outline: authenticate when retrieve metadata tagging result
        Given <exam_lo>add a exam_lo to topic
        And <exam_lo>create some tags
        And <exam_lo>add some quizzes to exam_lo with tags
        And <exam_lo>create study plan with book
        And <exam_lo>a student join course
        And <exam_lo>a signed in "student"
        And <exam_lo>a student do exam lo
        And <exam_lo>a signed in "<role>"
        Then <exam_lo>user retrieve metadata tagging result
        And <exam_lo>returns "<status>" status code

        Examples:
            | role         | status |
            | school admin | OK     |
            | admin        | OK     |
            | teacher      | OK     |
            | student      | OK     |
            | hq staff     | OK     |


    Scenario: retrieve metadata tagging result
        Given <exam_lo>add a exam_lo to topic
        And <exam_lo>create some tags
        And <exam_lo>add some quizzes to exam_lo with tags
        And <exam_lo>create study plan with book
        And <exam_lo>a student join course
        And <exam_lo>a signed in "student"
        And <exam_lo>a student do exam lo
        And <exam_lo>user retrieve metadata tagging result
        And <exam_lo>metadata tagging result is correct