Feature: Submit Assignment

    Background: A valid assignmewnt
        Given <assignment>a signed in "school admin"
        And <assignment>a valid book content


    Scenario: Student submit not-assigned assignment
        Given <assignment>a valid assignment
        And <assignment>a signed in "student"
        And a course and study plan with "no" student
        When student submit unrelated assignment
        Then <assignment>returns "PermissionDenied" status code

    Scenario Outline: Student submit assignment without any restriction
        Given <assignment>a valid assignment
        And <assignment>a signed in "student"
        And a course and study plan with "current" student
        When user submit their assignment "<times>" times
        Then our system must records all the submissions from student "<times>" times
        Examples:
            | times    |
            | single   |
            | multiple |

    Scenario Outline: Teacher submit task assignment without any restriction
        Given <assignment>a signed in "student"
        And a course and study plan with "current" student
        And <assignment>a valid task assignment
        And user submit their assignment "single" times
        And <assignment>a signed in "teacher"
        When user submit their assignment "<times>" times
        Then our system must records all the submissions from student "<times>" times
        Examples:
            | times    |
            | single   |
            | multiple |

    Scenario Outline: Teacher grade student submission
        Given <assignment>a valid assignment
        And <assignment>a signed in "student"
        And a course and study plan with "current" student
        And user submit their assignment with old submission endpoint
        And <assignment>a signed in "teacher"
        When teacher grade the assignments
        And our system must records highest grade from assignment
