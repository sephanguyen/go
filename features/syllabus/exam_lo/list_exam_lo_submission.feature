Feature: List exam lo submissions
    Background: a course in some valid locations
        Given <exam_lo>a signed in "teacher"
        And some students added to course in some valid locations

    Scenario: teacher listing exam lo submissions with locations
        Given a quiz test include "5" multiple choice quizzes with "5" quizzes per page and do quiz test for exam lo
        When student choose option "1, 2, 3" of the quiz "1" for submit quiz answers
        And student choose option "1, 2" of the quiz "2" for submit quiz answers
        And student choose option "1, 3" of the quiz "3" for submit quiz answers
        And student choose option "2, 3" of the quiz "4" for submit quiz answers
        And student choose option "1, 4" of the quiz "5" for submit quiz answers
        And student submit quiz answers
        And list exam lo submissions with valid locations
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submissions correctly


    Scenario: teacher listing exam lo submissions with locations
        Given create quiz tests and answers for exam lo
        And list exam lo submissions with valid locations
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submissions correctly

    Scenario: teacher listing exam lo submissions with invalid locations
        Given a quiz test include "5" multiple choice quizzes with "5" quizzes per page and do quiz test for exam lo
        When student choose option "1, 2, 3" of the quiz "1" for submit quiz answers
        And student choose option "1, 2" of the quiz "2" for submit quiz answers
        And student choose option "1, 3" of the quiz "3" for submit quiz answers
        And student choose option "2, 3" of the quiz "4" for submit quiz answers
        And student choose option "1, 4" of the quiz "5" for submit quiz answers
        And student submit quiz answers
        And list exam lo submissions with invalid locations
        Then <exam_lo>returns "OK" status code

    Scenario Outline: teacher listing exam lo submissions with filter
        Given all student answers and submit quizzes belong to exam
        And list exam lo submissions with filter by "<filter>"
        Then <exam_lo>returns "OK" status code
        And our system must returns list exam lo submissions with filter by "<filter>" correctly
        Examples:
            | filter                     |
            | student name               |
            | exam name                  |
            | random exam name           |
            | random student name        |
            | student name and exam name |
            | special character          |

    Scenario Outline: teacher listing exam lo submissions with filter and teacher manually grade exam submission
        Given all student answers and submit quizzes belong to exam
        And list exam lo submissions with filter by "<filter_1>"
        Then <exam_lo>returns "OK" status code
        Then teacher manually grade exam submission
        And list exam lo submissions with filter by "<filter_2>"
        And our system must returns list exam lo submissions with filter by "<filter_2>" correctly
        Examples:
            | filter_1     | filter_2  |
            | student name | corrector |


