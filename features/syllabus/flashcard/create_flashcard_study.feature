Feature: Create flashcard study

    Background:a valid book content
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And a valid flashcard with quizzes
        And <flashcard>a signed in "student"
        And <flashcard>a course and study plan with "current" student

    Scenario: Create flashcard study with roles
        Given <flashcard>a signed in "<role>"
        When user create flashcard study
        Then <flashcard>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | OK               |
            | school admin   | PermissionDenied |
            | hq staff       | PermissionDenied |
            | teacher        | PermissionDenied |
            # | centre lead    | PermissionDenied |
            # | centre manager | PermissionDenied |
            # | teacher lead   | PermissionDenied |

    Scenario: Student create flashcard study
        When user create flashcard study
        Then our system creates flashcard progression correctly
