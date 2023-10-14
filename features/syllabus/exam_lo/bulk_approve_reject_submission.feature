Feature: Bulk approve/reject submission

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

    Scenario Outline: authenticate when bulk approve/reject submission
        Given <exam_lo>a signed in "teacher"
        And user grades a submission answers to "SUBMISSION_STATUS_MARKED" status
        And <exam_lo>returns "OK" status code
        When <exam_lo>a signed in "<role>"
        And user action bulk "approve" submission
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | lead teacher   | PermissionDenied |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | OK               |
            | center staff   | PermissionDenied |

    Scenario Outline: HQ Staff action bulk approve/reject submission
        Given <exam_lo>a signed in "teacher"
        And user grades a submission answers to "<submission status>" status
        When <exam_lo>a signed in "hq staff"
        And user action bulk "<action>" submission
        Then <exam_lo>returns "<status code>" status code
        Then our system must returns "<status change>" status and "<result change>" result correctly

        Examples:
            | submission status          | action  | status code | status change                 | result change                        |
            | SUBMISSION_STATUS_MARKED   | approve | OK          | SUBMISSION_STATUS_RETURNED    | EXAM_LO_SUBMISSION_PASSED            |
            | SUBMISSION_STATUS_MARKED   | reject  | OK          | SUBMISSION_STATUS_IN_PROGRESS | EXAM_LO_SUBMISSION_WAITING_FOR_GRADE |
            | SUBMISSION_STATUS_RETURNED | reject  | OK          | SUBMISSION_STATUS_IN_PROGRESS | EXAM_LO_SUBMISSION_WAITING_FOR_GRADE |