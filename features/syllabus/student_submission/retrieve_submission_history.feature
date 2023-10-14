Feature: Retrieve submission by study plan item indentity

    Background: students submit assignments

    Scenario Outline: authenticate  retrive student submission history
        Given <student_submission> a signed in "school admin"
        And <student_submission> a valid book content with "<learning material>"
        And <student_submission> some students added to course in some valid locations
        And create a study plan for that course
        And student do test of "<learning material>"
        Given <student_submission> a signed in "<role>"
        When user retrieve student submission history

        Examples:
            | learning material  | role         | msg              |
            | learning objective | school admin | PermissionDenied |
            | learning objective | parent       | PermissionDenied |
            | learning objective | student      | PermissionDenied |
            | learning objective | teacher      | OK               |
            | learning objective | hq staff     | PermissionDenied |


    Scenario Outline: teacher retrive student submission history
        Given <student_submission> a signed in "school admin"
        And <student_submission> a valid book content with "<learning material>"
        And <student_submission> some students added to course in some valid locations
        And create a study plan for that course
        And student do test of "<learning material>"
        And <student_submission> a signed in "teacher"
        When user retrieve student submission history
        Then our system returns correct student submission history

        Examples:
            | learning material  |
            | learning objective |
            | flashcard          |