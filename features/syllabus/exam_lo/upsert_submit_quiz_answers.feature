Feature: Upsert submit quiz answers

    Background: a course in some valid locations
        Given <exam_lo>a signed in "school admin"
        And user creates a valid book content
        And user creates a course and add students into the course
        And user adds a master study plan with the created book

    Scenario Outline: authenticate when upsert submit quiz answers
        Given <exam_lo>a signed in "school admin"
        And user creates an Exam LO with manual grading is "true", grade to pass is "", approve grading is "false"
        And user adds 1 quizzes in "multiple choice" type and sets 1 point for each quiz
        And user updates study plan for the Exam LO
        Given <exam_lo>a signed in "<role>"
        And user starts and submits answers in multiple choice type
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | PermissionDenied |
            | admin          | PermissionDenied |
            | lead teacher   | PermissionDenied |
            | teacher        | PermissionDenied |
            | student        | OK               |
            | hq staff       | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario Outline: user upsert submit quiz answers
        Given <exam_lo>a signed in "school admin"
        And user creates an Exam LO with manual grading is "<manual grading>", grade to pass is "<grade to pass>", approve grading is "false"
        And user adds <number of quizzes> quizzes in "<types of quizzes>" type and sets <point> point for each quiz
        And user updates study plan for the Exam LO
        Given <exam_lo>a signed in "student"
        And user starts and submits answers in multiple choice type and exit
        And lo progression and lo progression answers has been created
        When user starts and submits answers in multiple choice type
        Then our system must return submit "<result>", <number of quizzes>, <point> correctly
        And lo progression and lo progression answers has been deleted correctly

        Examples:
            | manual grading | grade to pass | number of quizzes | types of quizzes | point | result                               |
            | true           | 5             | 5                 | multiple choice  | 2     | EXAM_LO_SUBMISSION_WAITING_FOR_GRADE |
            | false          |               | 5                 | multiple choice  | 2     | EXAM_LO_SUBMISSION_COMPLETED         |
            | false          | 2             | 5                 | multiple choice  | 2     | EXAM_LO_SUBMISSION_PASSED            |
            | false          | 11            | 5                 | multiple choice  | 2     | EXAM_LO_SUBMISSION_FAILED            |