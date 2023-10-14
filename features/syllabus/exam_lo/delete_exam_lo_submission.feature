Feature: Delete exam lo submission

    Background: a course in some valid locations
        Given <exam_lo>a signed in "school admin"
        And user creates a valid book content
        And user creates a course and add students into the course
        And user adds a master study plan with the created book
        And user creates an Exam LO with manual grading is "true", grade to pass is "5", approve grading is "false"
        And user adds 10 quizzes in "multiple choice" type and sets 1 point for each quiz
        And user updates study plan for the Exam LO
        Given <exam_lo>a signed in "student"
        And user starts and submits answers in multiple choice type
        And create student event logs after do quiz
        Given <exam_lo>a signed in "teacher"
        And user list exam lo submission scores
        And user grades a submission answers to "SUBMISSION_STATUS_IN_PROGRESS" status

    Scenario Outline: authenticate when delete exam lo submission
        Given <exam_lo>a signed in "<role>"
        When user delete exam lo submission
        Then <exam_lo>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | OK               |
            | center staff   | PermissionDenied |
            | lead teacher   | PermissionDenied |

    Scenario: hq staff delete exam lo submission
        Given <exam_lo>a signed in "hq staff"
        When user delete exam lo submission
        Then exam lo submission and related tables have been deleted correctly

