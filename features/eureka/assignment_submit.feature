Feature: Submit Assignment

    As a learner, I want to submit the answers for assigned homework

    Scenario: student submit not-assigned assignment
        Given some students
        And unrelated assignments
        When student submit random assignment
        Then our system must reject that

    Scenario Outline: Submit assignment without any restriction
        Given some students are assigned some valid study plans
        When student submit their "existed" content assignment "<times>" times
        Then our system must records all the submissions from student "<times>" times
        And all related study plan items mark as completed
        Examples:
            | times    |
            | single   |
            | multiple |

    Scenario: student must not allow listing submissions
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "student" list the submissions
        Then returns "PermissionDenied" status code

    Scenario: teacher listing submissions
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "teacher" list the submissions
        Then our system must returns only latest submission for each assignment

    Scenario: teacher listing submissions with invalid assignment name
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "teacher" list the submissions with invalid assignment name
        Then returns "OK" status code
        And our system must returns empty submission

    Scenario: teacher listing submissions with valid assignment name
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "teacher" list the submissions with assignment name
        Then returns "OK" status code
        And our system must returns only latest submission for each assignment
        And our system must returns submission with valid assignment name

    Scenario: teacher listing submissions with valid course id
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "teacher" list the submissions with course id
        Then returns "OK" status code
        And our system must returns only latest submission for each assignment
        And our system must returns submission with valid courses

    Scenario: student retrieve some else submissions
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "student" retrieve some else submissions
        Then returns "OK" status code

    Scenario: teacher grade submission
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When teacher grade the submissions multiple times with "exited" content
        Then our system must update the submissions with latest result

    Scenario: teacher change submission status
        Given some student has their submission graded
        When teacher change student's submission status to "SUBMISSION_STATUS_MARKED"
        Then our system must update the submissions status to "SUBMISSION_STATUS_MARKED"

    Scenario: student retrieve submission grade
        Given some student has their submission graded
        When "student" retrieve their submission grade
        Then returns "OK" status code

    Scenario: student retrieve submission grade
        Given some student has their submission graded
        And teacher change student's submission status to "SUBMISSION_STATUS_RETURNED"
        When "student" retrieve their submission grade
        Then returns "OK" status code

    Scenario: created at will change when submit multiple times
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "teacher" list the submissions
        Then our system must update created_at for each latest submission

    Scenario: teacher submit assignment
        Given some students are assigned some valid study plans
        When teacher submit "existed" content assignment "single" times
        Then our system must records all the submissions from student "single" times
        And all related study plan items mark as completed

    Scenario: student submit task assignment
        Given some students are assigned some valid study plans
        When student submit their "existed" content assignment "multiple" times
        Then our system must records all the submissions from student "multiple" times
        And our system must update daily learning time correctly
        And all related study plan items mark as completed
    
    @quarantined
    Scenario: teacher change status submissions when the teacher haven't graded
        Given some student has their submission haven't graded
        When teacher change student's submission status to "SUBMISSION_STATUS_RETURNED"
        Then our system must update the submissions status to "SUBMISSION_STATUS_RETURNED"
        And grade infomations have to included to submissions

    @quarantined
    Scenario: student submit their submission with no content
        Given some students are assigned some valid study plans
        When student submit their "none" content assignment "single" times for different assignments
        Then our system must update null content for each submission

    @quarantined
    Scenario: teacher grade submission without gradecontent
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "single" times for different assignments
        When teacher grade the submissions multiple times with "none" content
        Then our system must update the submissions with null content

    @quarantined
    Scenario: teacher change status submissions to returns which graded and not graded
        Given some student has their submission have graded and commented
        And some student has their submission haven't graded
        When teacher change student's submission status to "SUBMISSION_STATUS_RETURNED"
        Then our system must stores correctly

    @quarantined
    Scenario: teacher list submission after student is remove from a class
        Given student is remove from a class after they submit their submission
        When "teacher" list the submissions with course id
        Then returns "OK" status code
        And the response submissions don't contain submission of student who is removed from class

    @quarantined
    Scenario: teacher listing submissions with multi status filter
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        And modify not marked status to multi status
        And student submit their "existed" content assignment "single" times for different assignments
        When teacher list the submissions with status filter
        Then our system must returns only latest submission for each assignment
    
    @quarantined
    Scenario: student retrieve their own submissions
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        When "student" retrieve their own submissions
        Then our system must returns only latest submission for each assignment
