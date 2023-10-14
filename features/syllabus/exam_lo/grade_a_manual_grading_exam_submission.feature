Feature: Grade a manual grading exam submission

    Background: a list exam lo submission score
        Given <exam_lo>a signed in "school admin"
        And user creates a valid book content
        And user creates a course and add students into the course
        And user adds a master study plan with the created book
        And user creates an Exam LO with manual grading is "true", grade to pass is "5", approve grading is "false"
        And user adds 5 quizzes in "multiple choice" type and sets 2 point for each quiz
        And user updates study plan for the Exam LO
        Given <exam_lo>a signed in "student"
        And user starts and submits answers in multiple choice type
        Given <exam_lo>a signed in "teacher"
        And user list exam lo submission scores

    Scenario Outline: authenticate when grading a manual grading exam submission
        Given <exam_lo>a signed in "<role>"
        And user grades a submission answers to "SUBMISSION_STATUS_IN_PROGRESS" status
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | lead teacher   | OK               |
            | teacher        | OK               |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | OK               |
            | center staff   | PermissionDenied |

    Scenario Outline: teacher grades a manual grading exam submission
        Given <exam_lo>a signed in "teacher"
        And user grades a submission answers to "<submission status>" status
        And user grades a submission answers to "<status change>" status
        Then <exam_lo>returns "OK" status code
        And our system must returns graded score correctly

        Examples:
            | submission status            | status change                 |
            | SUBMISSION_STATUS_NOT_MARKED | SUBMISSION_STATUS_NOT_MARKED  |
            | SUBMISSION_STATUS_NOT_MARKED | SUBMISSION_STATUS_IN_PROGRESS |
            | SUBMISSION_STATUS_NOT_MARKED | SUBMISSION_STATUS_MARKED      |
            | SUBMISSION_STATUS_NOT_MARKED | SUBMISSION_STATUS_RETURNED    |