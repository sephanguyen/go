Feature: Retrieve submission by study plan item indentity

    Background: students submit assignments
        Given <student_submission> a signed in "school admin"
        And <student_submission> a valid book content
        And <student_submission> some students added to course in some valid locations
        And create a study plan for that course
        And students submit their assignments

    Scenario Outline: authenticate retrieve student submissions
        Given <student_submission> a signed in "<role>"
        When user retrieve submissions
        Then <student_submission> returns "<msg>" status code

        Examples:
            | role         | msg              |
            | school admin | OK               |
            | parent       | PermissionDenied |
            | student      | OK               |
            | teacher      | OK               |
            | hq staff     | OK               |
            # | centre lead    | PermissionDenied |
            # | centre manager | PermissionDenied |
            | lead teacher | OK               |

    Scenario: retrieve student submission
        Given <student_submission> a signed in "teacher"
        When user retrieve submissions
        Then retrieve student submissions correctly