Feature: List submissions v3
    #retrieve submissions of student mapped with those locations
    Background: students submit assignments
        Given <student_submission> a signed in "school admin"
        And <student_submission> a valid book content
        And <student_submission> some students added to course in some valid locations
        And create a study plan for that course
        And students submit their assignments

    Scenario Outline: authenticate  list submissions v3
        Given <student_submission> a signed in "<role>"
        When user using list submissions v3
        Then <student_submission> returns "<msg>" status code

        Examples:
            | role           | msg              |
            | school admin   | OK               |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | teacher        | OK               |
            | hq staff       | OK               |

    Scenario: student list submissions v3
        Given <student_submission> a signed in "teacher"
        When user using list submissions v3
        Then returns list student submissions correctly
